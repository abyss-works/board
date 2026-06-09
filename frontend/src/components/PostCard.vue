<script setup lang="ts">
import type { Post } from '@/types'

defineProps<{ post: Post }>()
defineEmits<{ click: [] }>()
</script>

<template>
  <div
    @click="$emit('click')"
    class="py-4 cursor-pointer hover:bg-gray-50 -mx-2 px-2 rounded-md transition-colors"
  >
    <h3 class="font-medium text-gray-900">{{ post.title }}</h3>
    <p class="text-sm text-gray-500 mt-1">
      {{ post.author }} · {{ new Date(post.createdAt).toLocaleDateString('ko-KR') }}
    </p>
  </div>
</template>
