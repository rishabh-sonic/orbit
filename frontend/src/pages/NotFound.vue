<template>
  <div class="flex flex-col items-center justify-center min-h-[calc(100vh-8rem)] gap-4 text-center">
    <h1 class="text-6xl font-bold text-muted-foreground/30">404</h1>
    <p class="text-xl font-semibold">Page not found</p>
    <p class="text-muted-foreground">The page you're looking for doesn't exist.</p>
    <Button as-child variant="outline">
      <RouterLink to="/">Go home</RouterLink>
    </Button>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button'
</script>
