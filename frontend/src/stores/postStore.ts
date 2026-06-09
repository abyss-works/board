import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchPosts as apiFetchPosts, createPost as apiCreatePost } from '@/api/posts'
import type { Post, CreatePostRequest } from '@/types'

export const usePostStore = defineStore('post', () => {
  const posts = ref<Post[]>([])
  const loading = ref(false)

  async function loadPosts() {
    loading.value = true
    try {
      posts.value = await apiFetchPosts()
    } finally {
      loading.value = false
    }
  }

  async function addPost(req: CreatePostRequest) {
    const post = await apiCreatePost(req)
    posts.value.unshift(post)
    return post
  }

  return { posts, loading, loadPosts, addPost }
})
