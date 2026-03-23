package worker

import (
	"context"
	"time"

	"github.com/luckysxx/user-platform/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// OutboxWorker 发件箱异步补偿 Worker
// 职责：定时轮询 event_outbox 表中 status=pending 的记录，逐条投递至 Kafka，
// 根据投递结果更新状态为 success 或累加 retry_count。
type OutboxWorker struct {
	outboxRepo  repository.EventOutboxRepository
	writer      *kafka.Writer
	redisClient *redis.Client
	logger      *zap.Logger

	interval  time.Duration // 轮询间隔
	batchSize int           // 每次拉取的最大条数
}

// NewOutboxWorker 构造函数
func NewOutboxWorker(
	outboxRepo repository.EventOutboxRepository,
	writer *kafka.Writer,
	redisClient *redis.Client,
	logger *zap.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo:  outboxRepo,
		writer:      writer,
		redisClient: redisClient,
		logger:      logger,
		interval:    1 * time.Minute, // 默认 1 分钟轮询一次做兜底
		batchSize:   50,              // 默认每次最多拉 50 条
	}
}

// Run 启动轮询，直到 ctx 被取消（优雅停机）
func (w *OutboxWorker) Run(ctx context.Context) error {
	w.logger.Info("OutboxWorker 启动",
		zap.Duration("interval", w.interval),
		zap.Int("batch_size", w.batchSize),
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	wakeupChan := make(chan struct{}, 1)

	// 后台协程：阻塞监听 Redis 唤醒信号
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			// 阻塞等待 5 秒，如果没有收到就回来继续外层循环（允许在 context 取消时及时退出）
			res, err := w.redisClient.BLPop(ctx, 5*time.Second, "outbox_wakeup_list").Result()
			if err == nil && len(res) > 0 {
				select {
				case wakeupChan <- struct{}{}:
					// 发送成功，唤醒轮询
				default:
					// 通道已满（说明已经发送了唤醒信号还没处理），不阻塞
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("OutboxWorker 收到停机信号，正在退出")
			return nil
		case <-ticker.C:
			w.poll(ctx)
		case <-wakeupChan:
			w.poll(ctx)
		}
	}
}

// poll 执行一次轮询：拉取 -> 投递 -> 更新状态
func (w *OutboxWorker) poll(ctx context.Context) {
	events, err := w.outboxRepo.GetPendingEvents(ctx, w.batchSize)
	if err != nil {
		w.logger.Error("OutboxWorker 拉取待发送事件失败", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return // 没有待处理的信件，安静等待下一轮
	}

	var msgs []kafka.Message
	var ids []int
	for _, evt := range events {
		ids = append(ids, evt.ID)
	}

	// 预处理消息
	for _, evt := range events {
		// 构造 Kafka 消息
		msgs = append(msgs, kafka.Message{
			Topic: evt.Topic,
			Value: evt.Payload,
		})
	}

	// 将这一批投递至 Kafka
	if err := w.writer.WriteMessages(ctx, msgs...); err != nil {
		w.logger.Error("OutboxWorker 投递 Kafka 失败",
			zap.Error(err),
		)
		// 投递失败：累加重试次数
		if markErr := w.outboxRepo.MarkEventRetry(ctx, ids); markErr != nil {
			w.logger.Error("OutboxWorker 标记事件失败状态异常", zap.Ints("event_ids", ids), zap.Error(markErr))
		}
		return
	}

	if markErr := w.outboxRepo.MarkEventAsSuccess(ctx, ids); markErr != nil {
		w.logger.Error("OutboxWorker 标记事件成功状态异常", zap.Error(markErr))
	}

	w.logger.Info("OutboxWorker 成功投递事件")
}
