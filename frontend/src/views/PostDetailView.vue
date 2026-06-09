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
const author = ref('')
const deleting = ref(false)

const postId = Number(route.params.id)

onMounted(async () => {
  post.value = await fetchPost(postId)
  comments.value = await fetchComments(postId)
})

async function submitComment() {
  if (!newComment.value.trim()) return
  const comment = await createComment(postId, {
    content: newComment.value,
    author: author.value || '익명',
  })
  comments.value.push(comment)
  newComment.value = ''
}

async function handleDelete() {
  if (!confirm('정말 삭제하시겠습니까?')) return
  deleting.value = true
  await store.removePost(postId)
  router.push('/')
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
            @click="handleDelete"
            :disabled="deleting"
            class="px-3 py-1.5 text-xs border border-red-200 rounded-md text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
          >
            {{ deleting ? '삭제 중...' : '삭제' }}
          </button>
        </div>
      </div>
      <p class="text-sm text-gray-500 mb-8">
        {{ post.author }} · {{ new Date(post.createdAt).toLocaleDateString('ko-KR') }}
      </p>
      <div class="prose prose-gray max-w-none whitespace-pre-wrap">{{ post.content }}</div>
    </article>

    <section class="border-t border-gray-200 pt-8">
      <h2 class="text-sm font-medium text-gray-500 mb-6">댓글 {{ comments.length }}개</h2>

      <div v-if="comments.length === 0" class="text-gray-400 text-sm mb-6">
        아직 댓글이 없습니다.
      </div>

      <ul v-else class="space-y-4 mb-8">
        <li v-for="c in comments" :key="c.id">
          <CommentItem :comment="c" />
        </li>
      </ul>

      <div class="space-y-3">
        <input
          v-model="author"
          placeholder="이름 (선택)"
          class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300"
        />
        <textarea
          v-model="newComment"
          placeholder="댓글을 입력하세요"
          rows="3"
          class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300 resize-none"
        />
        <button
          @click="submitComment"
          class="px-4 py-2 bg-gray-900 text-white rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors"
        >
          댓글 작성
        </button>
      </div>
    </section>
  </div>
</template>
