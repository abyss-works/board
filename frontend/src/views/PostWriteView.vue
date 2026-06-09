<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { usePostStore } from '@/stores/postStore'

const router = useRouter()
const store = usePostStore()
const title = ref('')
const content = ref('')
const author = ref('')

async function submit() {
  if (!title.value.trim() || !content.value.trim()) return
  const post = await store.addPost({
    title: title.value,
    content: content.value,
    author: author.value || '익명',
  })
  router.push(`/posts/${post.id}`)
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

    <h1 class="text-2xl font-serif font-semibold mb-8">글 작성</h1>

    <form @submit.prevent="submit" class="space-y-5">
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
          rows="8"
          class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300 resize-none"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-500 mb-1.5">이름 (선택)</label>
        <input
          v-model="author"
          placeholder="익명"
          class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300"
        />
      </div>
      <div class="flex gap-3">
        <button
          type="submit"
          class="px-5 py-2.5 bg-gray-900 text-white rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors"
        >
          등록
        </button>
        <button
          type="button"
          @click="router.push('/')"
          class="px-5 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-50 transition-colors"
        >
          취소
        </button>
      </div>
    </form>
  </div>
</template>
