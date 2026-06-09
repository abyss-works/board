import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'post-list',
      component: () => import('@/views/PostListView.vue'),
    },
    {
      path: '/posts/:id',
      name: 'post-detail',
      component: () => import('@/views/PostDetailView.vue'),
    },
    {
      path: '/write',
      name: 'post-write',
      component: () => import('@/views/PostWriteView.vue'),
    },
  ],
})

export default router
