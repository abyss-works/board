# Frontend

> 프론트엔드 기술 스택, 계층 아키텍처, 라우트, 인증.

---

# 기술 스택

| 영역 | 기술 |
|------|------|
| 프레임워크 | Vue 3 (Composition API) |
| 언어 | TypeScript |
| 빌드 | Vite |
| 상태 관리 | Pinia |
| 라우터 | Vue Router 4 |
| HTTP 클라이언트 | Axios |
| 스타일링 | Tailwind CSS |

---

# 계층 아키텍처

## 4계층

모든 도메인 로직은 다음 4계층을 통해 흘러야 한다. 계층을 건너뛰거나 우회해서는 절대 안 된다.

```
View (페이지/컴포넌트)
  → Store (상태 관리 + 액션)
    → Service (에러 변환, 비즈니스 로직)
      → API (HTTP 호출)
```

## 각 계층의 책임

| 계층 | 책임 | 금지 |
|------|------|------|
| **API** | HTTP 호출만 수행. 서버 응답을 `ApiResult<T>`로 반환. | 비즈니스 로직, 에러 처리, 직접 DOM 조작 |
| **Service** | API 응답 가공, 에러 변환. 사용자 친화적 메시지 제공. | HTTP 직접 호출, 상태 관리 |
| **Store** | 전역 상태, 비동기 액션, 로딩/에러 상태 관리. | HTTP 직접 호출, localStorage 직접 접근 |
| **View** | UI 렌더링, 사용자 인터랙션 처리. | Store 우회, API/Service 직접 호출, 비즈니스 로직 |

---

# 디렉토리 구조

```
src/
  api/              # API 계층 — axios 인스턴스, 엔드포인트별 요청 함수
  services/         # Service 계층 — 응답 가공, 비즈니스 로직
  stores/           # Store 계층 — Pinia 스토어
  views/            # View 계층 — 페이지 컴포넌트
  components/       # 공통/재사용 컴포넌트
  router/           # Vue Router 설정, 내비게이션 가드
  types/            # TypeScript 타입 정의
  utils/            # 유틸리티 함수
```

---

# 라우트

## 설계 규칙

| 규칙 | 설명 |
|------|------|
| **소문자 + 케밥** | `/posts`, `/posts/:id` — camelCase, PascalCase 금지 |
| **계층 구조** | `/`로 리소스 계층 표현 |
| **명사 기반** | URL은 리소스 명사로 구성 |

## 라우트 목록

| 경로 | 컴포넌트 | 설명 |
|------|----------|------|
| `/` | `PostListView` | 게시글 목록 |
| `/posts/:id` | `PostDetailView` | 게시글 상세 + 댓글 |
| `/write` | `PostWriteView` | 게시글 작성 |

## 내비게이션 가드

- `requiresAuth`: 미인증 시 로그인 페이지로 리다이렉트 (추후 도입)
- `guestOnly`: 인증 상태면 메인 페이지로 리다이렉트 (추후 도입)

---

# 인증 흐름

## 로그인 (추후 도입)

```
LoginView → authStore.login(credentials)
  → authService.login(request)
    → authApi.login(request)       // POST /api/auth/login
  ← { token: { accessToken }, user: { ... } }
  → accessToken, user → 상태 저장 → localStorage
```

## 인증 상태 유지

- 페이지 새로고침: `localStorage`에서 토큰/사용자 정보 복원
- API 호출: 공통 헤더 함수로 토큰 첨부
- 라우트 가드: 메타 필드로 접근 제어

## 로그아웃

```
authStore.logout()
  → accessToken = null, user = null
  → localStorage.removeItem(...)
```

---

# 토큰 스토리지

localStorage 키는 한 곳에서 중앙 관리한다:

| 키 | 용도 |
|----|------|
| `board-access-token` | 액세스 토큰 |
| `board-auth-user` | 로그인 사용자 정보 (JSON) |

---

# API 응답 형식

백엔드 Go API는 다음 JSON 형식으로 응답한다:

```json
{
  "id": 1,
  "title": "string",
  "content": "string",
  "author": "string",
  "created_at": "2024-01-01T00:00:00Z"
}
```

프론트엔드 `api/` 계층에서는 camelCase로 변환하여 사용한다 (`created_at` → `createdAt`).

에러 응답은 HTTP 상태 코드를 그대로 전파하며, 서비스 계층에서 사용자 친화적 메시지로 변환한다.
