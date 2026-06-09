import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchPosts as apiFetchPosts, createPost as apiCreatePost, updatePost as apiUpdatePost, deletePost as apiDeletePost } from '@/api/posts'
import type { Post, CreatePostRequest, UpdatePostRequest } from '@/types'

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

  async function updatePost(id: number, req: UpdatePostRequest) {
    const updated = await apiUpdatePost(id, req)
    const idx = posts.value.findIndex(p => p.id === id)
    if (idx !== -1) posts.value[idx] = updated
    return updated
  }

  async function removePost(id: number) {
    await apiDeletePost(id)
    posts.value = posts.value.filter(p => p.id !== id)
  }

  return { posts, loading, loadPosts, addPost, updatePost, removePost }
})
