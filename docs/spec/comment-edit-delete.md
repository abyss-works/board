# board-004: 댓글 수정/삭제 기능

## 이슈

- **Issue #10**: https://github.com/abyss-works/board/issues/10
- **기능 브랜치**: `feat/comment-edit-delete`
- **작성자**: PM
- **상태**: Draft

---

## 1. 기능 개요

### 배경

현재 게시글(`posts`)은 수정/삭제 기능이 구현되어 있으나, 댓글(`comments`)은 작성과 조회만 가능하다.
사용자가 댓글을 잘못 작성했을 때 수정하거나, 댓글을 삭제할 수 있는 기능이 필요하다.

### 목표

- 댓글 작성 시 비밀번호(선택)를 설정할 수 있도록 확장
- 댓글 수정(PUT) 및 삭제(DELETE) API 제공
- 비밀번호 검증을 통한 댓글 소유권 확인 (게시글 패턴과 동일)
- 프론트엔드에 댓글 수정/삭제 UI 추가

### 기존 게시글 수정/삭제 패턴 (참조)

게시글 수정/삭제는 `handlePostByID`의 PUT/DELETE 케이스에 이미 구현되어 있으며,
다음 패턴을 댓글에도 동일하게 적용한다:

| 항목 | 게시글 패턴 | 댓글 적용 방안 |
|------|-----------|-------------|
| 비밀번호 저장 | `password_hash TEXT` 컬럼, bcrypt 해싱 | 동일 |
| 수정 검증 | bcrypt.CompareHashAndPassword() | 동일 |
| 삭제 방식 | Soft delete (`deleted_at` IS NULL) | 동일 |
| 수정 시간 | `updated_at TIMESTAMP` 추적 | 동일 |
| 응답 코드 | 204 No Content (삭제), 200 (수정) | 동일 |

---

## 2. API 설계

### 2.1 엔드포인트 목록

| 메서드 | 경로 | 설명 | 변경 여부 |
|--------|------|------|----------|
| `GET` | `/api/posts/{postID}/comments` | 특정 게시글의 댓글 목록 조회 | 없음 |
| `POST` | `/api/posts/{postID}/comments` | 댓글 작성 | **확장**: `password` 필드 추가 |
| `GET` | `/api/comments?post_id={postID}` | 댓글 목록 조회 (쿼리 파라미터) | 없음 |
| `PUT` | `/api/comments/{commentID}` | 댓글 수정 | **신규** |
| `DELETE` | `/api/comments/{commentID}` | 댓글 삭제 | **신규** |

### 2.2 POST /api/posts/{postID}/comments — 댓글 작성 (확장)

**변경 사항**: 요청 바디에 `password` 필드 선택적 추가. 서버는 password_hash를 bcrypt로 생성하여 저장.

#### 요청

```json
{
  "content": "댓글 내용",
  "password": "1234"          // 선택 사항 (미입력 시 비밀번호 없음)
}
```

#### 응답 (201 Created)

```json
{
  "id": 1,
  "post_id": 42,
  "content": "댓글 내용",
  "author": "나그네",
  "created_at": "2026-06-09T12:00:00Z"
}
```

> ⚠️ 응답에 `password_hash`는 포함하지 않는다.

### 2.3 PUT /api/comments/{commentID} — 댓글 수정 (신규)

#### 요청

```json
{
  "content": "수정된 댓글 내용",
  "password": "1234"          // 필수 — 댓글 작성 시 설정한 비밀번호
}
```

#### 검증 흐름

```
POST /api/comments/{id}
  → DB에서 comment 조회 (deleted_at IS NULL)
    → 없음: 404 Not Found
    → password_hash가 NULL: 400 Bad Request ("비밀번호가 설정되지 않은 댓글입니다")
    → bcrypt.CompareHashAndPassword(password_hash, req.Password) 실패: 401 Invalid password
    → 성공: UPDATE content, updated_at
```

#### 응답 (200 OK)

```json
{
  "id": 1,
  "post_id": 42,
  "content": "수정된 댓글 내용",
  "author": "나그네",
  "created_at": "2026-06-09T12:00:00Z",
  "updated_at": "2026-06-09T13:00:00Z"
}
```

#### 에러 응답

| HTTP 코드 | 응답 바디 | 조건 |
|-----------|----------|------|
| 400 | `Comment ID required` | commentID 파싱 실패 |
| 400 | `Content required` | content가 비어 있음 |
| 400 | `Password required` | password가 비어 있음 |
| 400 | `Password not set for this comment` | 댓글에 password_hash가 없음 |
| 401 | `Invalid password` | 비밀번호 불일치 |
| 404 | `Comment not found` | commentID가 존재하지 않거나 삭제됨 |

### 2.4 DELETE /api/comments/{commentID} — 댓글 삭제 (신규)

#### 요청

```json
{
  "password": "1234"          // 필수 — 댓글 작성 시 설정한 비밀번호
}
```

#### 검증 흐름

```
DELETE /api/comments/{id}
  → DB에서 comment 조회 (deleted_at IS NULL)
    → 없음: 404 Not Found
    → password_hash가 NULL: 400 Bad Request
    → bcrypt.CompareHashAndPassword(password_hash, req.Password) 실패: 401 Invalid password
    → 성공: UPDATE deleted_at=NOW() (soft delete)
```

#### 응답 (204 No Content)

바디 없음.

#### 에러 응답

| HTTP 코드 | 응답 바디 | 조건 |
|-----------|----------|------|
| 400 | `Comment ID required` | commentID 파싱 실패 |
| 400 | `Password required` | password가 비어 있음 |
| 400 | `Password not set for this comment` | 댓글에 password_hash가 없음 |
| 401 | `Invalid password` | 비밀번호 불일치 |
| 404 | `Comment not found` | commentID가 존재하지 않거나 삭제됨 |

### 2.5 라우팅 구조

```go
// main.go main() 함수 내
http.HandleFunc("/api/posts", handlePosts)
http.HandleFunc("/api/posts/", handlePostByID)      // 기존: 개별 게시글 + 댓글 목록/작성
http.HandleFunc("/api/comments", handleComments)      // 기존: GET ?post_id=
http.HandleFunc("/api/comments/", handleCommentByID)  // 신규: PUT/DELETE 개별 댓글
```

`handleCommentByID`는 게시글의 `handlePostByID`와 동일한 패턴으로 구현:

```go
func handleCommentByID(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/comments/")
    parts := strings.Split(path, "/")
    commentID, err := strconv.Atoi(parts[0])
    // ... PUT/DELETE 분기
}
```

---

## 3. DB 마이그레이션

### 3.1 comments 테이블 변경 사항

기존 comments 테이블:

```sql
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT 'Anonymous',
    created_at TIMESTAMP DEFAULT NOW()
);
```

**추가할 컬럼**:

| 컬럼 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `password_hash` | TEXT | NULL | bcrypt 해시. NULL 허용 (비밀번호 없는 기존 댓글 호환) |
| `updated_at` | TIMESTAMP | NULL | 마지막 수정 시각 |
| `deleted_at` | TIMESTAMP | NULL | Soft delete 시각. NOT NULL이면 삭제된 댓글 |

### 3.2 마이그레이션 SQL (initDB에 추가)

```go
// 기존 테이블에 컬럼이 없으면 추가 (마이그레이션)
// -- posts 컬럼 (기존) --
db.Exec(`ALTER TABLE posts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP`)
db.Exec(`ALTER TABLE posts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`)
db.Exec(`ALTER TABLE posts ADD COLUMN IF NOT EXISTS password_hash TEXT`)

// -- comments 컬럼 (신규) --
db.Exec(`ALTER TABLE comments ADD COLUMN IF NOT EXISTS password_hash TEXT`)
db.Exec(`ALTER TABLE comments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP`)
db.Exec(`ALTER TABLE comments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`)
```

### 3.3 Comment 구조체 변경

```go
// main.go
type Comment struct {
    ID           int        `json:"id"`
    PostID       int        `json:"post_id"`
    Content      string     `json:"content"`
    Author       string     `json:"author"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    *time.Time `json:"updated_at,omitempty"`     // 신규
}
```

> `password_hash`와 `deleted_at`은 API 응답에 노출하지 않으므로 구조체에서 제외.
> DB 조회 시 필요한 경우 별도 변수로 처리.

### 3.4 기존 SELECT 쿼리 영향

기존 댓글 조회 쿼리에는 `deleted_at IS NULL` 조건을 추가하여 soft-delete된 댓글을 제외한다:

```go
// handlePostComments GET 수정
rows, err := db.Query(
    "SELECT id, post_id, content, author, created_at FROM comments WHERE post_id=$1 AND deleted_at IS NULL ORDER BY created_at ASC",
    postID,
)
```

---

## 4. 프론트엔드 변경사항

### 4.1 타입 정의 (types/index.ts)

```typescript
// Comment 인터페이스 확장
export interface Comment {
  id: number
  postId: number
  content: string
  author: string
  createdAt: string
  updatedAt?: string    // 신규
}

// CreateCommentRequest 확장
export interface CreateCommentRequest {
  content: string
  password?: string     // 신규 (선택)
}

// 신규 타입
export interface UpdateCommentRequest {
  content: string
  password: string
}

export interface DeleteCommentRequest {
  password: string
}
```

### 4.2 API 레이어 (api/comments.ts)

```typescript
import api from './index'
import type { Comment, CreateCommentRequest, UpdateCommentRequest, DeleteCommentRequest } from '@/types'

// 기존
export async function fetchComments(postId: number): Promise<Comment[]> { ... }
export async function createComment(postId: number, req: CreateCommentRequest): Promise<Comment> { ... }

// 신규
export async function updateComment(commentId: number, req: UpdateCommentRequest): Promise<Comment> {
  const { data } = await api.put<Comment>(`/comments/${commentId}`, req)
  return data
}

export async function deleteComment(commentId: number, password: string): Promise<void> {
  await api.delete(`/comments/${commentId}`, { data: { password } })
}
```

### 4.3 CommentItem.vue 컴포넌트 변경

댓글 항목에 수정/삭제 UI를 추가한다.

#### 상태

- `isEditing: boolean` — 수정 모드 여부
- `editContent: string` — 수정 중인 내용
- `editPassword: string` — 수정용 비밀번호
- `showDeletePrompt: boolean` — 삭제 확인 표시 여부
- `deletePassword: string` — 삭제용 비밀번호
- `error: string` — 에러 메시지

#### Props

```typescript
defineProps<{
  comment: Comment
}>()
```

#### Emits

```typescript
const emit = defineEmits<{
  (e: 'updated', comment: Comment): void
  (e: 'deleted', commentId: number): void
}>()
```

#### UI 레이아웃

```
┌──────────────────────────────────────────────────┐
│ 나그네 · 2026-06-09                     [수정] [삭제] │
│ 댓글 내용...                                       │
│                                                    │
│  ── 수정 모드 (isEditing=true) ──                  │
│  ┌──────────────────────────────────────┐          │
│  │ [textarea: 수정 내용]                  │          │
│  │ [input: 비밀번호]                      │          │
│  │ [저장] [취소]                          │          │
│  │ (에러 메시지)                          │          │
│  └──────────────────────────────────────┘          │
│                                                    │
│  ── 삭제 확인 (showDeletePrompt=true) ──            │
│  ┌──────────────────────────────────────┐          │
│  │ "댓글을 삭제하려면 비밀번호를 입력하세요"  │          │
│  │ [input: 비밀번호] [확인] [취소]          │          │
│  │ (에러 메시지)                          │          │
│  └──────────────────────────────────────┘          │
└──────────────────────────────────────────────────┘
```

### 4.4 PostDetailView.vue 변경

댓글 작성 폼에 비밀번호 필드(선택)를 추가한다.

```vue
<!-- 댓글 작성 영역 -->
<div class="space-y-3">
  <textarea v-model="newComment" ... />
  <div class="flex gap-2 items-end">
    <input
      v-model="newCommentPassword"
      type="password"
      placeholder="비밀번호 (수정/삭제 시 필요)"
      class="..."
    />
    <button @click="submitComment" ...>댓글 작성</button>
  </div>
</div>
```

`submitComment` 함수 수정:

```typescript
async function submitComment() {
  if (!newComment.value.trim()) return
  const comment = await createComment(postId, {
    content: newComment.value,
    password: newCommentPassword.value || undefined,
  })
  comments.value.push(comment)
  newComment.value = ''
  newCommentPassword.value = ''
}
```

CommentItem 이벤트 핸들링:

```typescript
function onCommentUpdated(updated: Comment) {
  const idx = comments.value.findIndex(c => c.id === updated.id)
  if (idx !== -1) comments.value[idx] = updated
}

function onCommentDeleted(commentId: number) {
  comments.value = comments.value.filter(c => c.id !== commentId)
}
```

```vue
<CommentItem
  :comment="c"
  @updated="onCommentUpdated"
  @deleted="onCommentDeleted"
/>
```

---

## 5. 작업 단위 분해 (Task Breakdown)

### 백엔드 태스크

| ID | 태스크 | 파일 | 설명 |
|----|--------|------|------|
| **BE-1** | comments 테이블 마이그레이션 | `main.go:initDB()` | `password_hash`, `updated_at`, `deleted_at` 컬럼 추가 (`ALTER TABLE ADD COLUMN IF NOT EXISTS`) |
| **BE-2** | Comment 구조체 확장 | `main.go` | `UpdatedAt *time.Time` 필드 추가 (`json:"updated_at,omitempty"`) |
| **BE-3** | 댓글 작성 핸들러 확장 | `main.go:handlePostComments POST` | 요청 바디에 `password` 필드 추가, bcrypt 해싱 후 `password_hash` 저장 |
| **BE-4** | 댓글 조회에 soft-delete 필터 추가 | `main.go:handlePostComments GET` | `WHERE deleted_at IS NULL` 조건 추가 |
| **BE-5** | `handleCommentByID` 함수 구현 | `main.go` (신규) | 개별 댓글 PUT/DELETE 핸들러 — 비밀번호 검증, 수정, soft-delete |
| **BE-6** | 라우트 등록 | `main.go:main()` | `http.HandleFunc("/api/comments/", handleCommentByID)` 추가 |
| **BE-7** | 기존 handleComments에 PUT/DELETE 차단 | `main.go:handleComments` | GET 외 메서드는 405 반환 확인 |

### 프론트엔드 태스크

| ID | 태스크 | 파일 | 설명 |
|----|--------|------|------|
| **FE-1** | 타입 정의 확장 | `frontend/src/types/index.ts` | `Comment.updatedAt`, `CreateCommentRequest.password`, `UpdateCommentRequest`, `DeleteCommentRequest` 추가 |
| **FE-2** | API 함수 추가 | `frontend/src/api/comments.ts` | `updateComment()`, `deleteComment()` 함수 구현 |
| **FE-3** | CommentItem 수정/삭제 UI | `frontend/src/components/CommentItem.vue` | 수정 버튼, 삭제 버튼, 인라인 수정 폼, 삭제 확인 다이얼로그, 이벤트 emit |
| **FE-4** | PostDetailView 댓글 작성 확장 | `frontend/src/views/PostDetailView.vue` | 비밀번호 입력 필드 추가, CommentItem 이벤트 핸들링 (`@updated`, `@deleted`) |

### 커밋 순서

1. **BE-1, BE-2** → Data layer: 테이블 마이그레이션 + 구조체 (`feat: 댓글 테이블에 password_hash, updated_at, deleted_at 컬럼 추가`)
2. **BE-3, BE-4** → 기존 핸들러 확장 (`feat: 댓글 작성 시 비밀번호 저장 및 soft-delete 필터 적용`)
3. **BE-5, BE-6, BE-7** → 신규 핸들러 + 라우팅 (`feat: 댓글 수정/삭제 API 구현`)
4. **FE-1, FE-2** → 타입 + API (`feat: 댓글 수정/삭제 API 레이어 추가`)
5. **FE-3, FE-4** → UI (`feat: 댓글 수정/삭제 UI 구현`)

---

## 6. 보안 고려사항

### 6.1 비밀번호 검증 로직

게시글과 동일한 bcrypt 기반 검증 패턴을 적용한다:

```go
// 비밀번호 검증 (게시글 PUT과 동일 패턴)
var hash sql.NullString
db.QueryRow(
    "SELECT password_hash FROM comments WHERE id=$1 AND deleted_at IS NULL",
    commentID,
).Scan(&hash)

if !hash.Valid {
    http.Error(w, "Password not set for this comment", 400)
    return
}
if bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(req.Password)) != nil {
    http.Error(w, "Invalid password", 401)
    return
}
```

### 6.2 보안 요구사항

| ID | 요구사항 | 설명 |
|----|---------|------|
| **SEC-1** | 비밀번호 해싱 | bcrypt (DefaultCost=10) 사용. 평문 저장 금지 |
| **SEC-2** | 비밀번호 미응답 | API 응답에 `password_hash` 필드 절대 포함하지 않음 |
| **SEC-3** | 비밀번호 없는 댓글 보호 | `password_hash`가 NULL인 댓글은 수정/삭제 불가 → 400 응답 |
| **SEC-4** | Soft delete | 실제 DELETE가 아닌 `deleted_at` 설정. 데이터 복구 가능성 확보 |
| **SEC-5** | SQL Injection 방지 | 모든 쿼리에 `$1`, `$2` 파라미터화된 쿼리 사용 (기존 패턴 유지) |
| **SEC-6** | Rate limiting 고려 | 현재 미구현. 추후 Issue로 등록 검토 (무차별 대입 공격 방어) |
| **SEC-7** | XSS 방지 | 프론트엔드에서 기본 이스케이프 처리 (Vue 템플릿 `{{ }}` 사용). 컨텐츠 렌더링 시 별도 조치 불필요 |

### 6.3 비밀번호 없는 기존 댓글 처리

기존 댓글(마이그레이션 이전)은 `password_hash`가 NULL이다.
이 댓글들에 대한 정책:

- **수정/삭제 불가**: 비밀번호가 없으므로 수정/삭제 시도 시 `400 Password not set for this comment` 반환
- **조회는 정상**: GET 응답에서는 기존과 동일하게 표시

> 추후 관리자 기능 도입 시 관리자에 의한 강제 삭제를 고려할 수 있으나, 이번 스펙 범위에서는 제외한다.

---

## 7. 영향도 분석

### 7.1 하위 호환성

| 영역 | 영향 | 대응 |
|------|------|------|
| 기존 API 응답 | Comment 구조체에 `updated_at` 필드 추가 (`omitempty`) | 기존 클라이언트는 무시 가능 |
| GET 댓글 목록 | `deleted_at IS NULL` 필터 추가 | 삭제된 댓글만 제외, 기존 동작과 동일 |
| POST 댓글 작성 | `password` 필드 선택적 추가 | 기존 클라이언트가 password 없이 호출해도 정상 동작 |
| DB 스키마 | 컬럼 3개 추가 (NULL 허용) | 기존 데이터 영향 없음 |

### 7.2 테스트 포인트

- [ ] 비밀번호 있는 댓글 작성 → 수정 성공
- [ ] 비밀번호 있는 댓글 작성 → 삭제 성공
- [ ] 잘못된 비밀번호로 수정 시도 → 401
- [ ] 잘못된 비밀번호로 삭제 시도 → 401
- [ ] 비밀번호 없는 댓글 수정 시도 → 400
- [ ] 비밀번호 없는 댓글 삭제 시도 → 400
- [ ] 삭제된 댓글이 GET 목록에서 제외됨
- [ ] 존재하지 않는 commentID → 404
- [ ] `updated_at`이 수정 후 정상 반영됨

### 7.3 변경 파일 요약

| 파일 | 작업 | 설명 |
|------|------|------|
| `main.go` | **수정** | initDB 마이그레이션, Comment 구조체, handlePostComments 확장, handleCommentByID 신규, 라우트 등록 |
| `frontend/src/types/index.ts` | **수정** | Comment, CreateCommentRequest 확장, UpdateCommentRequest, DeleteCommentRequest 신규 |
| `frontend/src/api/comments.ts` | **수정** | updateComment, deleteComment 함수 추가 |
| `frontend/src/components/CommentItem.vue` | **수정** | 수정/삭제 UI 추가, emit 이벤트 |
| `frontend/src/views/PostDetailView.vue` | **수정** | 비밀번호 입력 필드, CommentItem 이벤트 핸들링 |

---

## 8. 참고 자료

- Issue #10: https://github.com/abyss-works/board/issues/10
- 기존 게시글 수정/삭제 패턴: `main.go` `handlePostByID` PUT (L238-278), DELETE (L280-310)
- 게시글 비밀번호 검증 패턴: `bcrypt.CompareHashAndPassword` (L256, L291)
- Spec 001: `docs/spec/board-001-date-fix.md`
- Spec 003: `docs/spec/board-003-content-type.md`
