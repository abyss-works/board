<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchPost } from '@/api/posts'
import { fetchComments, createComment } from '@/api/comments'
import { usePostStore } from '@/stores/postStore'
import type { Post, Comment } from '@/types'
import CommentItem from '@/components/CommentItem.vue'

const route = useRoute()
const router = useRouter()
const store = usePostStore()
const post = ref<Post | null>(null)
const comments = ref<Comment[]>([])
const newComment = ref('')
const newCommentPassword = ref('')
const deleting = ref(false)
const deletePassword = ref('')
const showDeletePrompt = ref(false)
const deleteError = ref('')

const postId = Number(route.params.id)

onMounted(async () => {
  post.value = await fetchPost(postId)
  comments.value = await fetchComments(postId)
})

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

function onCommentUpdated(updated: Comment) {
  const idx = comments.value.findIndex(c => c.id === updated.id)
  if (idx !== -1) comments.value[idx] = updated
}

function onCommentDeleted(commentId: number) {
  comments.value = comments.value.filter(c => c.id !== commentId)
}

async function handleDelete() {
  if (!deletePassword.value) return
  deleting.value = true
  deleteError.value = ''
  try {
    await store.removePost(postId, deletePassword.value)
    router.push('/')
  } catch {
    deleteError.value = '비밀번호가 일치하지 않습니다'
    deleting.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 py-12">
    <button
      @click="router.push('/')"
      class="text-sm text-gray-500 hover:text-gray-900 mb-6 inline-block transition-colors"
    >
      ← 목록으로
    </button>

    <article v-if="post" class="mb-12">
      <div class="flex items-start justify-between gap-4 mb-3">
        <h1 class="text-2xl font-serif font-semibold">{{ post.title }}</h1>
        <div class="flex gap-2 shrink-0">
          <button
            @click="router.push(`/posts/${postId}/edit`)"
            class="px-3 py-1.5 text-xs border border-gray-200 rounded-md text-gray-600 hover:bg-gray-50 transition-colors"
          >
            수정
          </button>
          <button
            @click="showDeletePrompt = true"
            class="px-3 py-1.5 text-xs border border-red-200 rounded-md text-red-500 hover:bg-red-50 transition-colors"
          >
            삭제
          </button>
        </div>
      </div>
      <p class="text-sm text-gray-500 mb-8">
        나그네 · {{ new Date(post.createdAt).toLocaleDateString('ko-KR') }}
      </p>
      <div class="prose prose-gray max-w-none whitespace-pre-wrap">{{ post.content }}</div>
    </article>

    <!-- 삭제 비밀번호 입력 -->
    <div v-if="showDeletePrompt" class="mb-8 p-4 bg-red-50 border border-red-200 rounded-md">
      <p class="text-sm font-medium text-red-700 mb-2">글을 삭제하려면 비밀번호를 입력하세요</p>
      <div class="flex gap-2">
        <input
          v-model="deletePassword"
          type="password"
          placeholder="비밀번호"
          class="flex-1 px-3 py-2 bg-white rounded-md text-sm border border-red-200 focus:outline-none focus:ring-2 focus:ring-red-300"
        />
        <button
          @click="handleDelete"
          :disabled="deleting"
          class="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
        >
          {{ deleting ? '삭제 중...' : '확인' }}
        </button>
        <button
          @click="showDeletePrompt = false; deletePassword = ''; deleteError = ''"
          class="px-4 py-2 border border-gray-200 rounded-lg text-sm text-gray-600 hover:bg-gray-50 transition-colors"
        >
          취소
        </button>
      </div>
      <p v-if="deleteError" class="text-red-600 text-xs mt-1">{{ deleteError }}</p>
    </div>

    <section class="border-t border-gray-200 pt-8">
      <h2 class="text-sm font-medium text-gray-500 mb-6">댓글 {{ comments.length }}개</h2>

      <div v-if="comments.length === 0" class="text-gray-400 text-sm mb-6">
        아직 댓글이 없습니다.
      </div>

      <ul v-else class="space-y-4 mb-8">
        <li v-for="c in comments" :key="c.id">
          <CommentItem
            :comment="c"
            @updated="onCommentUpdated"
            @deleted="onCommentDeleted"
          />
        </li>
      </ul>

      <div class="space-y-3">
        <textarea
          v-model="newComment"
          placeholder="댓글을 입력하세요"
          rows="3"
          class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300 resize-none"
        />
        <div class="flex gap-2 items-end">
          <input
            v-model="newCommentPassword"
            type="password"
            placeholder="비밀번호 (수정/삭제 시 필요)"
            class="flex-1 px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300"
          />
          <button
            @click="submitComment"
            class="px-4 py-2 bg-gray-900 text-white rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors shrink-0"
          >
            댓글 작성
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
