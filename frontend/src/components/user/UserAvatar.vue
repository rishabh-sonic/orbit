<script setup lang="ts">
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'

const props = defineProps<{
  src?: string | null
  username: string
  size?: 'sm' | 'md' | 'lg'
}>()

const sizeClass: Record<string, string> = {
  sm: 'h-6 w-6 text-xs',
  md: 'h-8 w-8 text-sm',
  lg: 'h-10 w-10 text-base',
}
</script>

<template>
  <Avatar :class="sizeClass[size ?? 'md']">
    <AvatarImage v-if="src" :src="src" :alt="username" />
    <AvatarFallback>{{ username?.[0]?.toUpperCase() ?? '?' }}</AvatarFallback>
  </Avatar>
</template>
