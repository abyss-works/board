<script setup lang="ts">
import type { Comment } from '@/types'

defineProps<{ comment: Comment }>()
</script>

<template>
  <div class="text-sm">
    <p class="text-gray-500 mb-0.5">
      <span class="font-medium text-gray-700">{{ comment.author }}</span>
      <span class="mx-1.5 text-gray-300">·</span>
      {{ new Date(comment.createdAt).toLocaleDateString('ko-KR') }}
    </p>
    <p class="text-gray-800">{{ comment.content }}</p>
  </div>
</template>
