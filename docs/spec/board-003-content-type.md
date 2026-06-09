# board-003: Go 백엔드 JSON 응답에 Content-Type 헤더 누락 수정

## 배경 (Background)

### 이전 PR #2 실패 분석 (Issue #1 → PR #2)

Issue #1 ("Invalid Date" 버그)은 백엔드 `snake_case` ↔ 프론트엔드 `camelCase` 불일치 문제로 진단되었다.
PR #2에서는 **frontend axios interceptor** 로 해결을 시도했다:

1. `transform.ts` — snake_case ↔ camelCase 변환 유틸 (`snakeToCamel`, `camelToSnake`) 구현
2. `index.ts` — axios response interceptor에서 `Content-Type: application/json` 응답에 대해 `snakeToCamel()` 호출

**그러나 실제로는 동작하지 않았다.** 이유:

```
프론트엔드 interceptor 조건:
  if (response.headers['content-type']?.includes('application/json'))

Go 백엔드 GET 핸들러 실제 응답:
  Content-Type: text/plain; charset=utf-8   ← Go가 JSON 바디 첫 512바이트를 스니핑하여 자동 설정

결과:
  includes('application/json') → false → 변환 스킵 → Invalid Date 그대로 발생
```

### PM 스펙(001)이 이 문제를 놓친 이유

board-001 스펙은 **프론트엔드 interceptor 설계**에만 집중했다.
백엔드 응답의 실제 HTTP 헤더(`Content-Type`)를 분석 범위에 포함하지 않았다.
"axios interceptor를 추가하면 해결된다"는 가정이 틀렸고, **Go `net/http`의 Content-Type 자동 감지 동작**을 간과했다.

### Issue #3 개요

- **제목**: Go 백엔드 JSON 응답에 Content-Type 헤더 누락
- **원인**: GET 핸들러 3곳에서 `w.Header().Set("Content-Type", "application/json")`을 호출하지 않음
- **영향**: 프론트엔드 interceptor가 변환을 스킵 → Issue #1 근본 해결 안 됨
- **URL**: https://github.com/abyss-works/board/issues/3

---

## 요구사항 (Requirements)

### 기능 요구사항

| ID | 요구사항 | 설명 |
|----|---------|------|
| **FR-1** | 모든 JSON 응답에 `Content-Type: application/json` 헤더 설정 | GET/POST 모든 핸들러에서 JSON 응답 전에 Content-Type 명시적으로 설정 |
| **FR-2** | 프론트엔드 interceptor가 정상 동작 | 응답 헤더가 `application/json`이므로 `snakeToCamel()` 변환이 실행되어야 함 |
| **FR-3** | `curl -D-`로 Content-Type 헤더 검증 가능 | 모든 JSON 엔드포인트에서 `Content-Type: application/json` 확인 가능 |

### 비기능 요구사항

| ID | 요구사항 | 설명 |
|----|---------|------|
| **NFR-1** | 기존 POST 핸들러 동작 유지 | POST 핸들러(line 167, 249)는 이미 Content-Type 설정 중 — 변경하지 않음 |
| **NFR-2** | 에러 응답은 text/plain 유지 | `http.Error()` 응답은 기존대로 `text/plain; charset=utf-8` — JSON 아님 |
| **NFR-3** | 성능 영향 없음 | `w.Header().Set()`은 O(1) 해시맵 연산 |

---

## 기술 설계 (Technical Design)

### 전체 HTTP 요청/응답 흐름 분석

```
┌─────────────────────────────────────────────────────────────────────┐
│                        FULL HTTP FLOW                               │
│                                                                     │
│  ┌──────────┐    GET /api/posts         ┌──────────────────┐       │
│  │ Browser  │ ──────────────────────────│  Go net/http     │       │
│  │ (axios)  │                           │                  │       │
│  │          │ ◄── Response ─────────────│  handlePosts()   │       │
│  │          │    Content-Type: ???      │  GET (line 130)  │       │
│  └────┬─────┘                           └──────────────────┘       │
│       │                                                             │
│       │  axios interceptor (index.ts:10-14)                         │
│       │  ┌─────────────────────────────────────────────┐           │
│       │  │ if (content-type includes 'application/json')│           │
│       │  │   then snakeToCamel(data)                   │           │
│       │  │ else SKIP                                   │           │
│       │  └─────────────────────────────────────────────┘           │
│       │                                                             │
│       ▼                                                             │
│  ┌──────────┐                                                       │
│  │ Vue View │  post.createdAt === ???                               │
│  │ Component│  → undefined if interceptor skipped → Invalid Date    │
│  └──────────┘                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

### Go `net/http` Content-Type 자동 감지 동작

Go의 `net/http` 패키지는 `ResponseWriter.Write()` 호출 시 Content-Type이 설정되지 않았으면
**첫 512바이트를 `http.DetectContentType()`로 스니핑**하여 자동 설정한다.

```go
// Go 표준 라이브러리 동작 (net/http/sniff.go)
// JSON 바디 "[" 또는 "{" 로 시작 → "text/plain; charset=utf-8" 반환
// → Go는 JSON을 MIME 타입으로 인식하지 않음
```

`json.NewEncoder(w).Encode(data)`는 내부적으로 `w.Write()`를 호출하므로,
명시적 Content-Type 설정 없이 호출하면 `Content-Type: text/plain; charset=utf-8`이 된다.

### Go 백엔드 핸들러 전체 분석

| 핸들러 함수 | HTTP Method | line | JSON 응답? | Content-Type 설정? | 상태 |
|------------|-------------|------|-----------|-------------------|------|
| `handlePosts` | GET | 150 | ✅ `json.NewEncoder(w).Encode(posts)` | ❌ **없음** | 수정 필요 |
| `handlePosts` | POST | 167-169 | ✅ `json.NewEncoder(w).Encode(p)` | ✅ line 167 | 정상 |
| `handlePostByID` | GET | 204 | ✅ `json.NewEncoder(w).Encode(p)` | ❌ **없음** | 수정 필요 |
| `handlePostComments` | GET | 231 | ✅ `json.NewEncoder(w).Encode(comments)` | ❌ **없음** | 수정 필요 |
| `handlePostComments` | POST | 249-251 | ✅ `json.NewEncoder(w).Encode(c)` | ✅ line 249 | 정상 |
| `handleComments` | GET | 270 | 위임 → `handlePostComments` GET | ❌ **없음** | 위임 대상에서 수정 |
| `spaHandler` | GET | - | ❌ 정적 파일 | N/A | 해당 없음 |

**에러 응답** (`http.Error()`): 모두 `text/plain; charset=utf-8` → JSON이 아니므로 정상.

### 프론트엔드 Axios Interceptor 분석

```typescript
// frontend/src/api/index.ts line 10-14
api.interceptors.response.use((response) => {
  // ⚠️ 이 조건: Content-Type 헤더에 'application/json'이 포함되어야만 변환 실행
  if (response.data && String(response.headers['content-type'] ?? '').includes('application/json')) {
    response.data = snakeToCamel(response.data)
  }
  return response
})
```

| 응답 Content-Type | interceptor 변환? | 결과 |
|-------------------|-------------------|------|
| `application/json` | ✅ 실행 | `created_at` → `createdAt` (정상) |
| `text/plain; charset=utf-8` | ❌ 스킵 | `created_at` 그대로 → `createdAt` undefined → Invalid Date |
| 없음 (빈 문자열) | ❌ 스킵 | 위와 동일 |

### 해결 방안: GET 핸들러에 Content-Type 헤더 추가

`json.NewEncoder(w).Encode()` 호출 직전에 `w.Header().Set("Content-Type", "application/json")` 추가.

#### 수정 위치 (main.go)

**1. handlePosts GET (line 150 전)**
```go
// 수정 전 (line 130-150)
case "GET":
    rows, err := db.Query(...)
    // ... rows 처리 ...
    json.NewEncoder(w).Encode(posts)  // ← Content-Type 없음

// 수정 후
case "GET":
    rows, err := db.Query(...)
    // ... rows 처리 ...
    w.Header().Set("Content-Type", "application/json")  // ← 추가 (line 149 → NEW line 150)
    json.NewEncoder(w).Encode(posts)                     // ← 기존 line 150 → line 151
```

**2. handlePostByID GET (line 204 전)**
```go
// 수정 전 (line 196-205)
if r.Method == "GET" {
    var p Post
    err := db.QueryRow(...).Scan(...)
    if err != nil { ... }
    json.NewEncoder(w).Encode(p)  // ← Content-Type 없음
    return
}

// 수정 후
if r.Method == "GET" {
    var p Post
    err := db.QueryRow(...).Scan(...)
    if err != nil { ... }
    w.Header().Set("Content-Type", "application/json")  // ← 추가
    json.NewEncoder(w).Encode(p)
    return
}
```

**3. handlePostComments GET (line 231 전)**
```go
// 수정 전 (line 212-231)
case "GET":
    rows, err := db.Query(...)
    // ... rows 처리 ...
    json.NewEncoder(w).Encode(comments)  // ← Content-Type 없음

// 수정 후
case "GET":
    rows, err := db.Query(...)
    // ... rows 처리 ...
    w.Header().Set("Content-Type", "application/json")  // ← 추가
    json.NewEncoder(w).Encode(comments)
```

#### 수정 후 데이터 흐름 (To-Be)

```
Backend JSON: { "created_at": "2026-01-01T00:00:00Z", ... }
  Content-Type: application/json     ← ✅ 명시적 설정
       │
       ▼ (axios response interceptor)
  includes('application/json') → true → snakeToCamel 실행
       │
       ▼
Frontend receives: { createdAt: "2026-01-01T00:00:00Z", ... }
       │
       ▼
post.createdAt === "2026-01-01T00:00:00Z"  ← 정상
new Date(...) → 정상 날짜 객체
```

### 변경 파일 목록

| 파일 | 작업 | 설명 |
|------|------|------|
| `main.go` | **수정** | 3개 GET 핸들러에 `w.Header().Set("Content-Type", "application/json")` 추가 |
| `frontend/src/api/index.ts` | 변경 없음 | interceptor 조건문은 그대로 유지 |
| `frontend/src/api/transform.ts` | 변경 없음 | 변환 함수 그대로 유지 |

### 설계 결정 근거

1. **백엔드 수정이 올바른 접근인 이유**: Content-Type 헤더는 서버의 책임이다. JSON을 응답하면서 `text/plain`을 반환하는 것은 HTTP 스펙 위반에 가깝다. 프론트엔드에서 `content-type` 검사를 제거하면 다른 non-JSON 응답까지 변환 시도하는 부작용이 생긴다.
2. **interceptor 조건 유지**: `includes('application/json')` 체크는 방어적 코딩으로서 정당하다. 에러 응답(`text/plain`)이나 정적 파일 응답을 변환하지 않도록 보호한다.

---

## 검증 기준 (Verification)

### V1: Content-Type 헤더 검증 (curl)

서버 실행 후 모든 JSON 엔드포인트에서 `Content-Type: application/json`을 확인:

```bash
# 서버 실행
cd /home/ubuntu/springboot-app/board
DATABASE_URL="postgres://..." go run . &
SERVER_PID=$!
sleep 3

# 1. GET /api/posts
echo "=== GET /api/posts ==="
curl -s -D- http://localhost:8080/api/posts | head -20
# 기대: Content-Type: application/json

# 2. GET /api/posts/{id}
echo "=== GET /api/posts/1 ==="
curl -s -D- http://localhost:8080/api/posts/1 | head -20
# 기대: Content-Type: application/json

# 3. GET /api/posts/{id}/comments
echo "=== GET /api/posts/1/comments ==="
curl -s -D- http://localhost:8080/api/posts/1/comments | head -20
# 기대: Content-Type: application/json

# 4. GET /api/comments?post_id=1
echo "=== GET /api/comments?post_id=1 ==="
curl -s -D- "http://localhost:8080/api/comments?post_id=1" | head -20
# 기대: Content-Type: application/json

# 5. POST /api/posts (이미 Content-Type 설정됨 — 회귀 검증)
echo "=== POST /api/posts ==="
curl -s -D- -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","content":"Hello"}' | head -20
# 기대: Content-Type: application/json

kill $SERVER_PID 2>/dev/null
```

### V2: 프론트엔드 interceptor 동작 검증

브라우저 DevTools Network 탭에서:
1. Response Headers에 `Content-Type: application/json` 확인
2. Console에서 `snakeToCamel` 변환 완료된 객체 확인 (`createdAt` 필드 존재)
3. 게시글 목록/상세/댓글 화면에서 날짜가 "Invalid Date"가 아닌 실제 날짜로 표시됨

### V3: 에러 응답 회귀 검증

```bash
# 존재하지 않는 게시글 → 404 (text/plain 그대로)
curl -s -D- http://localhost:8080/api/posts/99999 | head -10
# 기대: 404 + Content-Type: text/plain; charset=utf-8
```

---

## 태스크 분할 (Task Breakdown)

### T1: main.go — GET 핸들러 3곳에 Content-Type 헤더 추가
- **파일**: `main.go`
- **위치**:
  - `handlePosts` GET: `json.NewEncoder(w).Encode(posts)` 직전 (line 150 앞)
  - `handlePostByID` GET: `json.NewEncoder(w).Encode(p)` 직전 (line 204 앞)
  - `handlePostComments` GET: `json.NewEncoder(w).Encode(comments)` 직전 (line 231 앞)
- **변경**: `w.Header().Set("Content-Type", "application/json")` 한 줄씩 추가
- **커밋**: `fix: GET 핸들러 JSON 응답에 Content-Type 헤더 추가`

### T2: 검증
- curl로 모든 GET/POST 엔드포인트에 대해 `Content-Type: application/json` 확인
- 프론트엔드 빌드 후 Invalid Date 재현 여부 확인
- 404/500 에러 응답이 여전히 `text/plain`인지 확인

### T3: PR 생성 및 리뷰
- dev → feat/3-content-type 브랜치 또는 dev 직접 커밋
- PR 설명에 `curl -D-` 출력 결과 첨부

---

## 수정이 필요한 핸들러 전체 목록

| # | 핸들러 | 메서드 | 파일:라인 | 현재 상태 | 수정 내용 |
|---|--------|--------|----------|----------|----------|
| 1 | `handlePosts` | GET | `main.go:150` | Content-Type 없음 | `w.Header().Set("Content-Type", "application/json")` 추가 |
| 2 | `handlePostByID` | GET | `main.go:204` | Content-Type 없음 | `w.Header().Set("Content-Type", "application/json")` 추가 |
| 3 | `handlePostComments` | GET | `main.go:231` | Content-Type 없음 | `w.Header().Set("Content-Type", "application/json")` 추가 |

> **참고**: `handleComments` GET (line 258-271)은 `handlePostComments`로 위임하므로 별도 수정 불필요.
> POST 핸들러(`handlePosts` POST line 167, `handlePostComments` POST line 249)는 이미 Content-Type 설정 완료.

---

## 참고 자료

- Issue #3: https://github.com/abyss-works/board/issues/3
- Issue #1: https://github.com/abyss-works/board/issues/1
- PR #2: 이전 수정 (interceptor만, 실패)
- Spec 001: `docs/spec/board-001-date-fix.md`
- Go `net/http` 소스: `src/net/http/sniff.go` (`DetectContentType()` — JSON을 인식하지 않음)
- Go `net/http` 문서: https://pkg.go.dev/net/http#ResponseWriter (Write 호출 시 Content-Type 자동 감지)
