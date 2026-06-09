# board-001: 게시글/댓글 날짜 "Invalid Date" 표시 버그 수정

## 배경 (Background)

### 증상
게시글 목록, 상세 페이지, 댓글 컴포넌트에서 모든 날짜가 "Invalid Date"로 표시된다.

### 버그 리포팅
- GitHub Issue: [#1](https://github.com/abyss-works/board/issues/1)
- 레이블: `customer-request`
- 보고일: 2026-06-09

### 원인 분석 (Root Cause)

백엔드와 프론트엔드 간 JSON 필드 네이밍 컨벤션 불일치.

| 계층 | 필드명 | 예시 |
|------|--------|------|
| 백엔드 (Go struct tag) | snake_case | `json:"created_at"` |
| 프론트엔드 (TypeScript) | camelCase | `createdAt: string` |

백엔드 `main.go`의 모든 도메인 struct는 snake_case JSON 태그를 사용한다:

```go
// main.go:29 (Post struct)
CreatedAt time.Time `json:"created_at"`

// main.go:37 (Comment struct)
CreatedAt time.Time `json:"created_at"`
```

프론트엔드 `types/index.ts`의 인터페이스는 camelCase를 사용한다:

```typescript
export interface Post {
  id: number
  createdAt: string  // camelCase
}

export interface Comment {
  postId: number     // camelCase
  createdAt: string  // camelCase
}
```

프로젝트 컨벤션 문서(`agent/project/frontend.md`) line 139에는 다음과 같이 명시되어 있다:

> "프론트엔드 api/ 계층에서는 camelCase로 변환하여 사용한다 (created_at → createdAt)"

그러나 이 변환 로직이 **구현되지 않았다**. `frontend/src/api/index.ts`는 단순히 axios 인스턴스를 생성할 뿐, 응답/요청에 대한 snake_case ↔ camelCase 변환 interceptor가 존재하지 않는다.

### 영향받는 파일

| 파일 | 라인 | 영향 |
|------|------|------|
| `frontend/src/api/index.ts` | 전체 | axios interceptor 부재 — 변환 로직 누락 |
| `frontend/src/components/PostCard.vue` | 15 | `new Date(post.createdAt)` → `undefined` 입력 |
| `frontend/src/components/CommentItem.vue` | 12 | `new Date(comment.createdAt)` → `undefined` 입력 |
| `frontend/src/views/PostDetailView.vue` | 46 | `new Date(post.createdAt)` → `undefined` 입력 |

### 데이터 흐름 (As-Is)

```
Backend JSON: { "created_at": "2026-01-01T00:00:00Z", ... }
       │
       ▼ (axios, no interceptor)
Frontend receives: { created_at: "2026-...", ... }
       │
       ▼ (typed as Post, but actual keys mismatch)
post.createdAt === undefined  ← 타입은 맞지만 실제 키가 없음
       │
       ▼
new Date(undefined) → "Invalid Date"
```

---

## 요구사항 (Requirements)

### 기능 요구사항

1. **FR-1**: API 응답의 모든 snake_case 키를 camelCase로 자동 변환한다.
2. **FR-2**: API 요청 body의 camelCase 키를 snake_case로 자동 변환한다 (백엔드 호환).
3. **FR-3**: 변환은 `api/` 계층(axios interceptor)에서 수행하며, View/Component 코드는 수정하지 않는다.
4. **FR-4**: 게시글 목록, 상세, 댓글 모든 화면에서 날짜가 정상적으로 표시되어야 한다 (`YYYY.MM.DD.` 형식).

### 비기능 요구사항

1. **NFR-1**: 성능 — 변환 로직은 O(n) depth=1 수준으로, 중첩 객체까지 재귀 변환하지 않는다 (현재 API 응답에 중첩 객체 없음).
2. **NFR-2**: 기존 API 시그니처 변경 없음 — `api/posts.ts`, `api/comments.ts`의 함수 시그니처는 그대로 유지한다.
3. **NFR-3**: 빌드 크기 증가 최소화 — lodash 등 외부 라이브러리 없이 순수 유틸 함수로 구현한다.

---

## 기술 설계 (Technical Design)

### 해결 방안: Axios Interceptor + 유틸 함수

axios 인스턴스에 `transformResponse` / `transformRequest`를 설정하여 JSON 키를 변환한다.

### snake_case ↔ camelCase 변환 유틸

```typescript
// frontend/src/api/transform.ts (신규 파일)

/**
 * snake_case → camelCase
 * 예: "created_at" → "createdAt", "post_id" → "postId"
 */
function toCamelCase(str: string): string {
  return str.replace(/_([a-z])/g, (_, c) => c.toUpperCase())
}

/**
 * camelCase → snake_case
 * 예: "createdAt" → "created_at", "postId" → "post_id"
 */
function toSnakeCase(str: string): string {
  return str.replace(/[A-Z]/g, (c) => '_' + c.toLowerCase())
}

/**
 * 객체의 모든 1-depth 키를 변환 (재귀 X — depth=1)
 */
function transformKeys<T>(obj: T, converter: (key: string) => string): T {
  if (Array.isArray(obj)) {
    return obj.map((item) => transformKeys(item, converter)) as T
  }
  if (obj !== null && typeof obj === 'object') {
    const result: Record<string, unknown> = {}
    for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
      result[converter(key)] = value
    }
    return result as T
  }
  return obj
}

export function snakeToCamel<T>(obj: T): T {
  return transformKeys(obj, toCamelCase)
}

export function camelToSnake<T>(obj: T): T {
  return transformKeys(obj, toSnakeCase)
}
```

### Axios Interceptor 적용

```typescript
// frontend/src/api/index.ts (수정)

import axios from 'axios'
import { snakeToCamel, camelToSnake } from './transform'

const api = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

// 응답: snake_case → camelCase
api.interceptors.response.use((response) => {
  if (response.data && response.headers['content-type']?.includes('application/json')) {
    response.data = snakeToCamel(response.data)
  }
  return response
})

// 요청: camelCase → snake_case
api.interceptors.request.use((config) => {
  if (config.data && typeof config.data === 'object') {
    config.data = camelToSnake(config.data)
  }
  return config
})

export default api
```

### 데이터 흐름 (To-Be)

```
Backend JSON: { "created_at": "2026-01-01T00:00:00Z", ... }
       │
       ▼ (axios response interceptor → snakeToCamel)
Frontend receives: { createdAt: "2026-01-01T00:00:00Z", ... }
       │
       ▼ (Post type matches actual keys now)
post.createdAt === "2026-01-01T00:00:00Z"  ← 정상 매핑
       │
       ▼
new Date("2026-01-01T00:00:00Z") → "2026. 6. 1." (ko-KR)
```

### 파일 변경 목록

| 파일 | 작업 | 설명 |
|------|------|------|
| `frontend/src/api/transform.ts` | **신규** | snake_case ↔ camelCase 변환 유틸 |
| `frontend/src/api/index.ts` | **수정** | axios interceptor 추가 (request + response) |
| `frontend/src/components/PostCard.vue` | 변경 없음 | — |
| `frontend/src/components/CommentItem.vue` | 변경 없음 | — |
| `frontend/src/views/PostDetailView.vue` | 변경 없음 | — |

### 설계 결정 근거

1. **Interceptor 접근법 선택 이유**: View/Component 레이어를 전혀 건드리지 않고 문제를 해결할 수 있다. 프로젝트 컨벤션(`frontend.md`)에서 의도한 "api/ 계층 변환"을 정확히 구현하는 방식이다.
2. **Depth=1 제한**: 현재 API 응답에 중첩 객체가 없으므로 재귀 변환은 YAGNI. 필요 시 추후 확장 가능.
3. **외부 라이브러리 미사용**: `camelcase-keys`, `humps` 등 라이브러리를 도입하지 않고 순수 유틸로 구현하여 빌드 크기와 의존성을 최소화한다.

---

## 태스크 분할 (Task Breakdown)

### T1: 변환 유틸리티 구현
- **파일**: `frontend/src/api/transform.ts` (신규)
- **내용**: `snakeToCamel()`, `camelToSnake()` 함수 구현
- **검증**: 유닛 테스트 (추후 도입 시) 또는 수동 확인
- **커밋**: `feat: snake_case-camelCase 변환 유틸 함수 추가`

### T2: Axios interceptor 적용
- **파일**: `frontend/src/api/index.ts` (수정)
- **내용**: response/request interceptor에 변환 함수 연결
- **검증**: API 호출 후 응답 데이터의 키가 camelCase인지 확인
- **커밋**: `fix: api 계층 snake_case→camelCase 변환 적용`

### T3: 통합 검증
- **내용**:
  - 게시글 목록에서 날짜 정상 표시 확인
  - 게시글 상세에서 날짜 정상 표시 확인
  - 댓글에서 날짜 정상 표시 확인
  - 게시글/댓글 작성 시 요청 body가 snake_case로 전송되는지 확인
- **커밋**: (없음 — T1, T2 커밋으로 충분)

### 커밋 컨벤션

프로젝트 규칙에 따라:
- 타입: `feat:` (신규 유틸), `fix:` (버그 수정)
- 제목: 한글, 명사형 종결
- 계층별 분리: API 계층 2개 커밋으로 구성

---

## 참고 자료

- Issue: https://github.com/abyss-works/board/issues/1
- 프로젝트 컨벤션: `agent/project/frontend.md` line 139 (camelCase 변환 명시)
- 백엔드 struct: `main.go` lines 29, 37 (`json:"created_at"`)
- 프론트엔드 타입: `frontend/src/types/index.ts` (`createdAt: string`)
