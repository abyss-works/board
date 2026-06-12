<script setup lang="ts">
import { ref } from 'vue'
import { updateComment, deleteComment } from '@/api/comments'
import type { Comment } from '@/types'

const props = defineProps<{ comment: Comment }>()
const emit = defineEmits<{
  (e: 'updated', comment: Comment): void
  (e: 'deleted', commentId: number): void
}>()

const isEditing = ref(false)
const editContent = ref('')
const editPassword = ref('')
const editError = ref('')
const editing = ref(false)

const showDeletePrompt = ref(false)
const deletePassword = ref('')
const deleteError = ref('')
const deleting = ref(false)

function startEdit() {
  editContent.value = props.comment.content
  editPassword.value = ''
  editError.value = ''
  isEditing.value = true
  showDeletePrompt.value = false
}

function cancelEdit() {
  isEditing.value = false
  editError.value = ''
}

async function handleUpdate() {
  if (!editContent.value.trim()) return
  if (!editPassword.value) {
    editError.value = '비밀번호를 입력하세요'
    return
  }
  editing.value = true
  editError.value = ''
  try {
    const updated = await updateComment(props.comment.id, {
      content: editContent.value,
      password: editPassword.value,
    })
    emit('updated', updated)
    isEditing.value = false
  } catch {
    editError.value = '비밀번호가 일치하지 않습니다'
  } finally {
    editing.value = false
  }
}

async function handleDelete() {
  if (!deletePassword.value) {
    deleteError.value = '비밀번호를 입력하세요'
    return
  }
  deleting.value = true
  deleteError.value = ''
  try {
    await deleteComment(props.comment.id, deletePassword.value)
    emit('deleted', props.comment.id)
    showDeletePrompt.value = false
  } catch {
    deleteError.value = '비밀번호가 일치하지 않습니다'
    deleting.value = false
  }
}
</script>

<template>
  <div class="text-sm">
    <div class="flex items-start justify-between gap-2">
      <p class="text-gray-500 mb-0.5">
        <span class="font-medium text-gray-700">{{ comment.author }}</span>
        <span class="mx-1.5 text-gray-300">·</span>
        {{ new Date(comment.createdAt).toLocaleDateString('ko-KR') }}
        <span v-if="comment.updatedAt" class="ml-1 text-gray-400 text-xs">
          (수정됨)
        </span>
      </p>
      <div v-if="!isEditing && !showDeletePrompt" class="flex gap-1.5 shrink-0">
        <button
          @click="startEdit"
          class="px-2.5 py-1 text-xs border border-gray-200 rounded-md text-gray-600 hover:bg-gray-50 transition-colors"
        >
          수정
        </button>
        <button
          @click="showDeletePrompt = true; deletePassword = ''; deleteError = ''"
          class="px-2.5 py-1 text-xs border border-red-200 rounded-md text-red-500 hover:bg-red-50 transition-colors"
        >
          삭제
        </button>
      </div>
    </div>

    <!-- 댓글 내용 -->
    <p v-if="!isEditing" class="text-gray-800 whitespace-pre-wrap">{{ comment.content }}</p>

    <!-- 수정 모드 -->
    <div v-if="isEditing" class="mt-2 space-y-2">
      <textarea
        v-model="editContent"
        rows="3"
        class="w-full px-3 py-2 bg-gray-50 rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300 resize-none"
      />
      <div class="flex gap-2 items-start">
        <input
          v-model="editPassword"
          type="password"
          placeholder="비밀번호"
          class="flex-1 px-3 py-1.5 bg-white rounded-md text-sm border border-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300"
        />
        <button
          @click="handleUpdate"
          :disabled="editing"
          class="px-3 py-1.5 bg-gray-900 text-white rounded-md text-xs font-medium hover:bg-gray-800 transition-colors disabled:opacity-50"
        >
          {{ editing ? '저장 중...' : '저장' }}
        </button>
        <button
          @click="cancelEdit"
          class="px-3 py-1.5 border border-gray-200 rounded-md text-xs text-gray-600 hover:bg-gray-50 transition-colors"
        >
          취소
        </button>
      </div>
      <p v-if="editError" class="text-red-600 text-xs">{{ editError }}</p>
    </div>

    <!-- 삭제 확인 -->
    <div v-if="showDeletePrompt" class="mt-2 p-3 bg-red-50 border border-red-200 rounded-md">
      <p class="text-xs font-medium text-red-700 mb-2">댓글을 삭제하려면 비밀번호를 입력하세요</p>
      <div class="flex gap-2">
        <input
          v-model="deletePassword"
          type="password"
          placeholder="비밀번호"
          class="flex-1 px-3 py-1.5 bg-white rounded-md text-sm border border-red-200 focus:outline-none focus:ring-2 focus:ring-red-300"
        />
        <button
          @click="handleDelete"
          :disabled="deleting"
          class="px-3 py-1.5 bg-red-600 text-white rounded-md text-xs font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
        >
          {{ deleting ? '삭제 중...' : '확인' }}
        </button>
        <button
          @click="showDeletePrompt = false; deletePassword = ''; deleteError = ''"
          class="px-3 py-1.5 border border-gray-200 rounded-md text-xs text-gray-600 hover:bg-gray-50 transition-colors"
        >
          취소
        </button>
      </div>
      <p v-if="deleteError" class="text-red-600 text-xs mt-1">{{ deleteError }}</p>
    </div>
  </div>
</template>
