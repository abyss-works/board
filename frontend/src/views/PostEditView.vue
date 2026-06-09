<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchPost } from '@/api/posts'
import { usePostStore } from '@/stores/postStore'

const route = useRoute()
const router = useRouter()
const store = usePostStore()

const postId = Number(route.params.id)
const title = ref('')
const content = ref('')
const loading = ref(true)

onMounted(async () => {
  const post = await fetchPost(postId)
  title.value = post.title
  content.value = post.content
  loading.value = false
})

async function submit() {
  if (!title.value.trim() || !content.value.trim()) return
  await store.updatePost(postId, {
    title: title.value,
    content: content.value,
  })
  router.push(`/posts/${postId}`)
}
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 py-12">
    <button
      @click="router.push(`/posts/${postId}`)"
      class="text-sm text-gray-500 hover:text-gray-900 mb-6 inline-block transition-colors"
    >
      ← 뒤로
    </button>

    <h1 class="text-2xl font-serif font-semibold mb-8">글 수정</h1>

    <div v-if="loading" class="text-gray-400 text-sm py-8 text-center">
      불러오는 중...
    </div>

    <form v-else @submit.prevent="submit" class="space-y-5">
      <div>
        <label class="block text-sm font-medium text-gray-500 mb-1.5">제목</label>
        <input
          v-model="title"
          placeholder="제목을 입력하세요"
          class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-500 mb-1.5">내용</label>
        <textarea
          v-model="content"
          placeholder="내용을 입력하세요"
          rows="12"
          class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300 resize-none"
        />
      </div>
      <div class="flex gap-3">
        <button
          type="submit"
          class="px-5 py-2.5 bg-gray-900 text-white rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors"
        >
          저장
        </button>
        <button
          type="button"
          @click="router.push(`/posts/${postId}`)"
          class="px-5 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-50 transition-colors"
        >
          취소
        </button>
      </div>
    </form>
  </div>
</template>
