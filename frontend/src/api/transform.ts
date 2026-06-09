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
 * 배열인 경우 각 요소에 대해 재귀적으로 변환
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

/**
 * snake_case 객체 키 → camelCase 객체 키 변환
 */
export function snakeToCamel<T>(obj: T): T {
  return transformKeys(obj, toCamelCase)
}

/**
 * camelCase 객체 키 → snake_case 객체 키 변환
 */
export function camelToSnake<T>(obj: T): T {
  return transformKeys(obj, toSnakeCase)
}
