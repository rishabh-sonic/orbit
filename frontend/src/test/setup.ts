import { config } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, vi } from 'vitest'

// Provide a proper localStorage mock so Pinia stores can use it in jsdom
const localStorageStore: Record<string, string> = {}
const localStorageMock = {
  getItem: (key: string) => localStorageStore[key] ?? null,
  setItem: (key: string, value: string) => {
    localStorageStore[key] = value
  },
  removeItem: (key: string) => {
    delete localStorageStore[key]
  },
  clear: () => {
    Object.keys(localStorageStore).forEach((k) => delete localStorageStore[k])
  },
}
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
})

// Reset pinia and localStorage before each test
beforeEach(() => {
  localStorageMock.clear()
  setActivePinia(createPinia())
})

// Global stubs for RouterLink and RouterView so components don't need a router
config.global.stubs = {
  RouterLink: {
    template: '<a :href="String(to)"><slot /></a>',
    props: ['to'],
  },
  RouterView: { template: '<div />' },
}

// Stub window.fetch globally — tests override per-case
vi.stubGlobal('fetch', vi.fn())
