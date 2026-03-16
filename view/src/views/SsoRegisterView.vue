<script setup lang="ts">
import { reactive } from 'vue'

const registerForm = reactive({
  email: '',
  username: '',
  password: '',
})

const handleRegister = () => {
  window.alert(`注册演示: ${registerForm.username}`)
}
</script>

<template>
  <main class="register-page">
    <section class="surface">
      <header class="top-bar">
        <div class="top-left">
          <div class="logo-mark" aria-hidden="true">L</div>
          <strong class="logo-text">luckys</strong>
        </div>

        <div class="top-right">
          <button type="button" class="text-btn">简体中文</button>
          <a href="#" class="text-link">意见反馈</a>
        </div>
      </header>

      <div class="hero-bg" aria-hidden="true">
        <span class="sweep" />
        <span class="cloud cloud-back" />
        <span class="cloud cloud-mid" />
        <span class="cloud cloud-front" />
        <span class="flow flow-a" />
        <span class="flow flow-b" />
        <span class="flow flow-c" />
        <span class="flow flow-d" />
        <span class="mist mist-a" />
        <span class="mist mist-b" />
      </div>

      <section class="auth-card">
        <h1>欢迎注册 luckys</h1>
        <p class="sub-title">统一身份账号中心</p>

        <form class="auth-form" @submit.prevent="handleRegister">
          <label class="form-item">
            <input v-model="registerForm.email" type="email" placeholder="邮箱" required />
          </label>

          <label class="form-item">
            <input v-model="registerForm.username" type="text" placeholder="用户名" required />
          </label>

          <label class="form-item">
            <input v-model="registerForm.password" type="password" placeholder="密码" required />
          </label>

          <button type="submit" class="primary-btn">立即注册</button>
        </form>
      </section>
    </section>
  </main>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700;800&display=swap');

:global(*) {
  box-sizing: border-box;
}

:global(body) {
  margin: 0;
}

.register-page {
  --bg-top: #b7d8ea;
  --bg-bottom: #d9e6ef;
  --input-bg: rgba(245, 249, 252, 0.72);
  --input-line: rgba(173, 193, 209, 0.72);
  --text: #111827;
  --muted: #5f6f84;
  --brand: #1492ec;

  min-height: 100vh;
  padding: 0;
  color: var(--text);
  font-family: 'Outfit', 'Avenir Next', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  background: linear-gradient(180deg, var(--bg-top) 0%, var(--bg-bottom) 100%);
}

.surface {
  position: relative;
  width: 100%;
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 104px 16px 84px;
  overflow: hidden;
}

.top-bar {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 3;
  width: 100%;
  height: 68px;
  padding: 0 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(184, 214, 232, 0.58);
  border-bottom: 1px solid rgba(112, 151, 181, 0.14);
}

.top-bar::after {
  content: '';
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.58), transparent);
}

.top-left,
.top-right {
  display: flex;
  align-items: center;
}

.top-right {
  gap: 14px;
}

.logo-mark {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: grid;
  place-items: center;
  font-size: 1.05rem;
  font-weight: 700;
  color: #ffffff;
  background: #111827;
}

.logo-text {
  margin-left: 10px;
  font-size: 1.28rem;
  letter-spacing: 0.02em;
}

.text-btn {
  border: none;
  background: transparent;
  color: #1f2f43;
  font-size: 1rem;
  cursor: pointer;
  transition: color 0.2s ease;
}

.text-link {
  color: #1f2f43;
  font-size: 1rem;
  text-decoration: none;
  transition: color 0.2s ease;
}

.text-btn:hover,
.text-link:hover {
  color: #103b63;
}

.hero-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
  background:
    radial-gradient(58% 42% at 86% 20%, rgba(92, 183, 236, 0.52), transparent 74%),
    radial-gradient(120% 58% at 14% 40%, rgba(232, 245, 252, 0.62), transparent 60%),
    linear-gradient(180deg, rgba(170, 205, 227, 0.2), rgba(216, 229, 239, 0.08));
}

.hero-bg::before {
  content: '';
  position: absolute;
  inset: -8% -4% auto -4%;
  height: 46%;
  background:
    radial-gradient(110% 58% at 56% 10%, rgba(255, 255, 255, 0.42), transparent 66%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.26), transparent 72%);
  opacity: 0.92;
  animation: shimmer 16s ease-in-out infinite;
}

.hero-bg::after {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle at 1px 1px, rgba(181, 212, 231, 0.22) 1px, transparent 0);
  background-size: 22px 22px;
  opacity: 0.24;
}

.cloud {
  position: absolute;
  border-radius: 44% 56% 54% 46% / 56% 48% 52% 44%;
  background:
    radial-gradient(32% 34% at 26% 28%, rgba(255, 255, 255, 0.88), rgba(255, 255, 255, 0.08) 72%, transparent 100%),
    radial-gradient(30% 32% at 76% 30%, rgba(255, 255, 255, 0.78), rgba(255, 255, 255, 0.06) 74%, transparent 100%),
    radial-gradient(72% 68% at 50% 62%, rgba(236, 246, 252, 0.76), rgba(173, 209, 231, 0.28) 68%, rgba(132, 181, 213, 0.12) 92%, transparent 100%);
  box-shadow:
    inset 0 18px 30px rgba(255, 255, 255, 0.34),
    inset 0 -24px 30px rgba(137, 184, 213, 0.2),
    0 24px 42px rgba(118, 165, 196, 0.22);
  opacity: 0.8;
  will-change: transform;
}

.cloud-back {
  width: 44vw;
  height: 22vw;
  min-width: 520px;
  min-height: 250px;
  right: -10vw;
  top: -4vh;
  opacity: 0.62;
  animation: cloud-drift-a 30s ease-in-out infinite;
}

.cloud-mid {
  width: 58vw;
  height: 28vw;
  min-width: 720px;
  min-height: 340px;
  left: -22vw;
  top: 30vh;
  opacity: 0.74;
  animation: cloud-drift-b 34s ease-in-out infinite;
}

.cloud-front {
  width: 60vw;
  height: 34vw;
  min-width: 760px;
  min-height: 380px;
  left: 18vw;
  bottom: -30vh;
  opacity: 0.84;
  animation: cloud-drift-c 38s ease-in-out infinite;
}

.flow {
  position: absolute;
  border-radius: 50%;
  border: 1px solid rgba(178, 214, 236, 0.54);
  opacity: 0.9;
  box-shadow: 0 14px 30px rgba(120, 173, 204, 0.16);
  will-change: transform;
}

.flow-a {
  width: 122vw;
  height: 58vw;
  left: -66vw;
  top: 34vh;
  background: radial-gradient(ellipse at 62% 48%, rgba(255, 255, 255, 0.72), rgba(213, 234, 247, 0.44) 58%, rgba(161, 201, 225, 0.18) 82%, transparent 100%);
  animation: float-a 24s ease-in-out infinite;
}

.flow-b {
  width: 86vw;
  height: 46vw;
  right: -30vw;
  top: -12vh;
  background: radial-gradient(ellipse at 34% 52%, rgba(255, 255, 255, 0.68), rgba(200, 228, 244, 0.36) 62%, rgba(108, 181, 223, 0.2) 82%, transparent 100%);
  animation: float-b 28s ease-in-out infinite;
}

.flow-c {
  width: 82vw;
  height: 50vw;
  left: 16vw;
  bottom: -36vh;
  background: radial-gradient(ellipse at 44% 40%, rgba(255, 255, 255, 0.66), rgba(192, 222, 240, 0.34) 62%, rgba(108, 179, 221, 0.18) 84%, transparent 100%);
  animation: float-c 32s ease-in-out infinite;
}

.flow-d {
  width: 64vw;
  height: 34vw;
  left: 42vw;
  top: 50vh;
  background: radial-gradient(ellipse at 50% 46%, rgba(255, 255, 255, 0.6), rgba(196, 225, 242, 0.28) 62%, transparent 100%);
  opacity: 0.68;
  animation: float-d 26s ease-in-out infinite;
}

.mist {
  position: absolute;
  border-radius: 50%;
  filter: blur(14px);
  opacity: 0.5;
}

.mist-a {
  width: 32vw;
  height: 18vw;
  left: -6vw;
  top: 10vh;
  background: radial-gradient(ellipse at 52% 50%, rgba(255, 255, 255, 0.56), rgba(202, 227, 242, 0.28) 66%, transparent 100%);
  animation: drift-a 20s ease-in-out infinite;
}

.mist-b {
  width: 28vw;
  height: 16vw;
  right: 6vw;
  bottom: 14vh;
  background: radial-gradient(ellipse at 52% 50%, rgba(255, 255, 255, 0.52), rgba(188, 219, 239, 0.24) 66%, transparent 100%);
  animation: drift-b 24s ease-in-out infinite;
}

.sweep {
  position: absolute;
  top: -18vh;
  left: -38vw;
  width: 28vw;
  height: 140vh;
  transform: rotate(12deg);
  background: linear-gradient(
    90deg,
    rgba(255, 255, 255, 0) 0%,
    rgba(255, 255, 255, 0.08) 35%,
    rgba(255, 255, 255, 0.34) 50%,
    rgba(255, 255, 255, 0.1) 65%,
    rgba(255, 255, 255, 0) 100%
  );
  filter: blur(2px);
  opacity: 0.7;
  animation: sweep-pass 14s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

.auth-card {
  position: relative;
  z-index: 1;
  width: min(360px, 100%);
  transform: translateY(-24px);
  padding: 0;
}

.auth-card h1 {
  margin: 0;
  text-align: center;
  font-size: 2.0rem;
  font-weight: 700;
}

.sub-title {
  margin: 8px 0 0;
  text-align: center;
  font-size: 0.88rem;
  letter-spacing: 0.08em;
  color: rgba(38, 69, 95, 0.72);
}

.auth-form {
  margin-top: 18px;
  display: grid;
  gap: 16px;
}

.form-item {
  display: grid;
  gap: 0;
}

.auth-form input[type='text'],
.auth-form input[type='email'],
.auth-form input[type='password'] {
  width: 100%;
  height: 42px;
  border: 1px solid var(--input-line);
  border-radius: 6px;
  padding: 0 12px;
  color: #26364a;
  background: linear-gradient(180deg, rgba(250, 253, 255, 0.92), rgba(239, 246, 251, 0.74));
  outline: none;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.68);
}

.auth-form input::placeholder {
  color: #9fb0c0;
}

.auth-form input:focus {
  border-color: rgba(20, 146, 236, 0.65);
  background: rgba(250, 253, 255, 0.92);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.primary-btn {
  margin-top: 2px;
  height: 42px;
  border: none;
  border-radius: 6px;
  font-size: 1.05rem;
  font-weight: 600;
  color: #ffffff;
  background: linear-gradient(180deg, #2aa3f3, #1492ec);
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(20, 146, 236, 0.25);
}

.primary-btn:hover {
  background: linear-gradient(180deg, #2aa3f3, #1492ec);
  box-shadow: 0 8px 18px rgba(20, 146, 236, 0.25);
}

@keyframes float-a {
  0% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
  50% {
    transform: translate3d(2.2vw, -1vh, 0) rotate(2.5deg);
  }
  100% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
}

@keyframes float-b {
  0% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
  50% {
    transform: translate3d(-2vw, 1.2vh, 0) rotate(-2.2deg);
  }
  100% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
}

@keyframes float-c {
  0% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
  50% {
    transform: translate3d(-1.6vw, -1.1vh, 0) rotate(1.8deg);
  }
  100% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
}

@keyframes float-d {
  0% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
  50% {
    transform: translate3d(1.5vw, -1vh, 0) rotate(-1.6deg);
  }
  100% {
    transform: translate3d(0, 0, 0) rotate(0deg);
  }
}

@keyframes drift-a {
  0% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(1.4vw, -0.8vh, 0);
  }
  100% {
    transform: translate3d(0, 0, 0);
  }
}

@keyframes drift-b {
  0% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(-1.2vw, 0.7vh, 0);
  }
  100% {
    transform: translate3d(0, 0, 0);
  }
}

@keyframes shimmer {
  0% {
    transform: translate3d(0, 0, 0);
    opacity: 0.74;
  }
  50% {
    transform: translate3d(1vw, -0.5vh, 0);
    opacity: 0.92;
  }
  100% {
    transform: translate3d(0, 0, 0);
    opacity: 0.74;
  }
}

@keyframes sweep-pass {
  0% {
    transform: translate3d(0, 0, 0) rotate(12deg);
    opacity: 0;
  }
  12% {
    opacity: 0.68;
  }
  50% {
    transform: translate3d(88vw, 0, 0) rotate(12deg);
    opacity: 0.68;
  }
  70% {
    opacity: 0;
  }
  100% {
    transform: translate3d(132vw, 0, 0) rotate(12deg);
    opacity: 0;
  }
}

@keyframes cloud-drift-a {
  0% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(-1.6vw, 1vh, 0);
  }
  100% {
    transform: translate3d(0, 0, 0);
  }
}

@keyframes cloud-drift-b {
  0% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(1.8vw, -1vh, 0);
  }
  100% {
    transform: translate3d(0, 0, 0);
  }
}

@keyframes cloud-drift-c {
  0% {
    transform: translate3d(0, 0, 0);
  }
  50% {
    transform: translate3d(-1.4vw, -0.8vh, 0);
  }
  100% {
    transform: translate3d(0, 0, 0);
  }
}

@media (max-width: 980px) {
  .top-bar {
    padding: 0 16px;
  }

  .top-right {
    gap: 8px;
  }

  .text-btn,
  .text-link {
    font-size: 0.9rem;
  }

  .auth-card h1 {
    font-size: 2.1rem;
  }
}

@media (max-width: 620px) {
  .surface {
    padding: 92px 14px 74px;
  }

  .top-bar {
    height: 60px;
  }

  .logo-mark {
    width: 26px;
    height: 26px;
    border-radius: 7px;
    font-size: 0.92rem;
  }

  .logo-text {
    font-size: 1.1rem;
  }

  .text-link {
    display: none;
  }

  .auth-card h1 {
    font-size: 1.2rem;
  }

  .auth-card {
    transform: translateY(-12px);
  }
}
</style>
