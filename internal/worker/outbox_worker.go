package worker

import (
	"context"
	"time"

	"github.com/luckysxx/user-platform/internal/repository"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// OutboxWorker 发件箱异步补偿 Worker
// 职责：定时轮询 event_outbox 表中 status=pending 的记录，逐条投递至 Kafka，
// 根据投递结果更新状态为 success 或累加 retry_count。
type OutboxWorker struct {
	outboxRepo repository.EventOutboxRepository
	writer     *kafka.Writer
	logger     *zap.Logger

	interval  time.Duration // 轮询间隔
	batchSize int           // 每次拉取的最大条数
}

// NewOutboxWorker 构造函数
func NewOutboxWorker(
	outboxRepo repository.EventOutboxRepository,
	writer *kafka.Writer,
	logger *zap.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo: outboxRepo,
		writer:     writer,
		logger:     logger,
		interval:   3 * time.Second, // 默认 3 秒轮询一次
		batchSize:  50,              // 默认每次最多拉 50 条
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

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("OutboxWorker 收到停机信号，正在退出")
			return nil
		case <-ticker.C:
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

	w.logger.Info("OutboxWorker 拉取到待发送事件", zap.Int("count", len(events)))

	for _, evt := range events {
		// 构造 Kafka 消息
		msg := kafka.Message{
			Topic: evt.Topic,
			Value: evt.Payload,
		}

		// 尝试投递至 Kafka
		if err := w.writer.WriteMessages(ctx, msg); err != nil {
			w.logger.Error("OutboxWorker 投递 Kafka 失败",
				zap.Int("event_id", evt.ID),
				zap.String("topic", evt.Topic),
				zap.Error(err),
			)
			// 投递失败：累加重试次数（这里简单地标记为 failed，后续可以加 retry_count 逻辑）
			if markErr := w.outboxRepo.MarkEventAsFailed(ctx, evt.ID); markErr != nil {
				w.logger.Error("OutboxWorker 标记事件失败状态异常", zap.Int("event_id", evt.ID), zap.Error(markErr))
			}
			continue
		}

		// 投递成功：标记为 success
		if markErr := w.outboxRepo.MarkEventAsSuccess(ctx, evt.ID); markErr != nil {
			w.logger.Error("OutboxWorker 标记事件成功状态异常", zap.Int("event_id", evt.ID), zap.Error(markErr))
		}

		w.logger.Info("OutboxWorker 成功投递事件",
			zap.Int("event_id", evt.ID),
			zap.String("topic", evt.Topic),
		)
	}
}
