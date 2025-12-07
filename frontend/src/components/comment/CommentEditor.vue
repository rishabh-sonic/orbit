<script setup lang="ts">
import { ref } from 'vue'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  placeholder?: string
  loading?: boolean
  autofocus?: boolean
}>()

const emit = defineEmits<{
  submit: [content: string]
  cancel: []
}>()

const content = ref('')

function submit() {
  if (!content.value.trim()) return
  emit('submit', content.value)
  content.value = ''
}
</script>

<template>
  <div class="space-y-2">
    <Textarea
      v-model="content"
      :placeholder="placeholder ?? 'Write a comment…'"
      rows="3"
      :autofocus="autofocus"
      class="resize-none"
    />
    <div class="flex gap-2 justify-end">
      <Button type="button" variant="outline" size="sm" @click="emit('cancel')">Cancel</Button>
      <Button size="sm" :disabled="loading || !content.trim()" @click="submit">
        {{ loading ? 'Posting…' : 'Post' }}
      </Button>
    </div>
  </div>
</template>
