<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePostStore } from '@/stores/postStore'
import PostCard from '@/components/PostCard.vue'

const router = useRouter()
const store = usePostStore()

onMounted(() => store.loadPosts())
</script>

<template>
  <div class="max-w-2xl mx-auto px-4 py-12">
    <header class="mb-10">
      <h1 class="text-3xl font-serif font-semibold tracking-tight">커뮤니티</h1>
      <p class="mt-1 text-gray-500 text-sm">이야기를 나누는 공간</p>
    </header>

    <div class="mb-8">
      <button
        @click="router.push('/write')"
        class="px-5 py-2.5 bg-gray-900 text-white rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors"
      >
        글 작성하기
      </button>
    </div>

    <section>
      <h2 class="text-sm font-medium text-gray-500 mb-4 pb-2 border-b border-gray-200">
        게시글
        <span v-if="store.posts.length" class="ml-1 text-gray-400">{{ store.posts.length }}</span>
      </h2>

      <div v-if="store.loading" class="text-gray-400 text-sm py-8 text-center">
        불러오는 중...
      </div>

      <div v-else-if="store.posts.length === 0" class="text-gray-400 text-sm py-12 text-center">
        아직 게시글이 없습니다. 첫 글을 작성해보세요.
      </div>

      <ul v-else class="divide-y divide-gray-100">
        <li v-for="post in store.posts" :key="post.id">
          <PostCard :post="post" @click="router.push(`/posts/${post.id}`)" />
        </li>
      </ul>
    </section>
  </div>
</template>
