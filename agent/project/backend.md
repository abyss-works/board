# Backend

> 백엔드 기술 스택, 모듈 구조, 계층 아키텍처, API, 보안.

---

# 기술 스택

| 영역 | 기술 |
|------|------|
| 런타임 | Go 1.22 |
| 프레임워크 | `net/http` (표준 라이브러리) |
| 빌드 | `go build` |
| DB 드라이버 | `github.com/lib/pq` (PostgreSQL) |
| DB | PostgreSQL (운영/개발 공통) |
| 보안 | JWT (추후 도입) |
| API 문서 | 코드 주석 기반 (추후 Swagger 도입 검토) |

---

# 모듈 구조

현재 단일 `main.go` 파일에서 HTTP 핸들러, DB 접근, HTML 템플릿을 모두 처리한다.
지속적 개발을 위해 다음 구조로 점진적 분리한다:

```
board/
  main.go              # 진입점: DB 초기화, 라우트 등록, 서버 시작
  handler/             # HTTP 핸들러 — 요청 파싱, 응답 직렬화
    post.go
    comment.go
  service/             # 비즈니스 로직
    post.go
    comment.go
  repository/          # 데이터 접근 (SQL 쿼리)
    post.go
    comment.go
  model/               # 도메인 모델 (struct 정의)
    post.go
    comment.go
  middleware/           # CORS, 인증 필터 등
    auth.go
```

분리 전까지는 `main.go` 내 동일 패키지 함수로 계층을 구분한다.

---

# 계층 아키텍처

## 3계층 (현재)

```
Handler (HTTP 요청/응답 처리)
  → Repository (SQL 쿼리 직접 실행)
    → DB (PostgreSQL)
```

`main.go`에서 handler 함수가 DB 쿼리를 직접 실행하는 구조.

## 목표 4계층

```
Handler (HTTP 요청/응답)
  → Service (비즈니스 로직)
    → Repository (데이터 접근)
      → DB (PostgreSQL)
```

계층을 건너뛰거나 우회하는 것은 금지된다.

## 의존성 방향

```
Handler ──→ Service ←── Repository
               ↓
             Model
```

- Handler → Service. 역방향 금지.
- Repository → Service. 역방향 금지.
- Model은 모든 계층에서 참조 가능.
- Service는 외부 계층에 의존하지 않는다.

## 요청 처리 흐름

```
Client Request
  → Middleware (인증/인가 — 추후)
    → Handler (입력 검증, JSON 파싱)
      → Service (비즈니스 로직 — 추후)
        → Repository (SQL 쿼리)
          → PostgreSQL
        ← Model struct
      ← 처리 결과
    ← JSON 직렬화
```

---

# API 응답 형식

현재는 도메인 struct를 직접 JSON으로 직렬화한다.

```json
// GET /api/posts 응답 예시
[
  {
    "id": 1,
    "title": "첫 번째 글",
    "content": "안녕하세요",
    "author": "홍길동",
    "created_at": "2024-03-15T10:30:00Z"
  }
]
```

## ApiResult 도입 (추후)

모든 API 응답을 `ApiResult<T>`로 감싸는 방식을 검토:

| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `success` | `bool` | ✅ | 요청 성공 여부 |
| `data` | `T` | 성공 시 | 응답 payload. 실패 시 `null` |
| `error` | `ApiError` | 실패 시 | 에러 정보. 성공 시 `null` |

---

# API 설계

## URL 규칙

| 규칙 | 설명 |
|------|------|
| **복수형 명사** | `/api/posts`, `/api/comments` |
| **소문자 + 케밥** | camelCase, snake_case 금지 |
| **Restful API** | 리소스 조회/생성/수정/삭제는 HTTP 메서드로 표현 |

## HTTP 메서드

| 메서드 | 용도 |
|--------|------|
| `GET` | 목록/단건 조회 |
| `POST` | 신규 생성 |
| `PUT` | 전체 수정 |
| `PATCH` | 부분 수정, 상태 변경 |
| `DELETE` | 삭제 |

---

# 엔티티

## 현재 모델 (`main.go` struct)

```go
type Post struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Author    string    `json:"author"`
    CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
    ID        int       `json:"id"`
    PostID    int       `json:"post_id"`
    Content   string    `json:"content"`
    Author    string    `json:"author"`
    CreatedAt time.Time `json:"created_at"`
}
```

## 구조 원칙

| 원칙 | 설명 |
|------|------|
| **PK** | 컬럼명 `id`, `SERIAL` (auto-increment) |
| **JSON 태그** | snake_case 사용 (`json:"created_at"`) |
| **시간** | `time.Time`, DB는 `TIMESTAMP` |

---

# 보안

## 현재 상태

- 인증 없음. 모든 엔드포인트 공개.
- 작성자명은 사용자 입력값. Anonymous 기본값.

## 도입 예정

1. JWT 기반 인증 (Access Token + Refresh Token)
2. `Authorization: Bearer <token>` 헤더 검증
3. 미들웨어에서 토큰 검증 → Context에 사용자 정보 설정
4. 실패 시 401 / 403

---

# 엔드포인트 목록

### 게시글 (`/api/posts`)

| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | `/api/posts` | 게시글 목록 조회 (최신순) |
| POST | `/api/posts` | 게시글 작성 |
| GET | `/api/posts/{id}` | 게시글 단건 조회 |

### 댓글 (`/api/posts/{id}/comments`)

| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | `/api/posts/{id}/comments` | 댓글 목록 조회 (등록순) |
| POST | `/api/posts/{id}/comments` | 댓글 작성 |
| GET | `/api/comments?post_id={id}` | 댓글 조회 (대체 엔드포인트) |

### 정적 파일

| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | `/` | SPA 진입점 (현재 서버 사이드 HTML, 추후 분리) |
