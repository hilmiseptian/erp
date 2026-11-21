
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import SupplierListView from '@/views/suppliers/SupplierListView.vue'
import MainLayout from '@/components/layout/MainLayout.vue'
import ItemListView from '@/views/items/ItemListView.vue'
import PurchaseRequestView from '@/views/purchasing/PurchaseRequestView.vue'
import PurchaseOrderView from '@/views/purchasing/PurchaseOrderView.vue'
import GoodReceiptView from '@/views/purchasing/GoodReceiptView.vue'
import StockView from '@/views/inventory/StockView.vue'
import StockAdjustmentView from '@/views/inventory/StockAdjustmentView.vue'
import SwaggerView from '@/views/SwaggerView.vue'
import CustomerListView from '@/views/customers/CustomerListView.vue'
import SalesOrderView from '@/views/sales/SalesOrderView.vue'
import GoodDeliveryView from '@/views/inventory/GoodDeliveryView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: {
        guestOnly: true,
      },
    },

    {
      path: '/',
      component: MainLayout,
      meta: {
        requiresAuth: true,
      },
      children: [
        {
          path: '',
          redirect: '/dashboard',
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: DashboardView,
          meta: {
            title: 'Dashboard',
          },
        },
        {
          path: 'items',
          name: 'items',
          component: ItemListView,
          meta: { title: 'Master Item' },
        },
        {
          path: 'suppliers',
          name: 'suppliers',
          component: SupplierListView,
          meta: {
            title: 'Suppliers',
          },
        },
        {
          path: 'sales/customers',
          name: 'customers',
          component: CustomerListView,
          meta: { title: 'Customers' },
        },
        {
          path: 'sales/orders',
          name: 'sales-orders',
          component: SalesOrderView,
          meta: { title: 'Sales Orders' },
        },
        {
          path: 'inventory/purchase-requests',
          name: 'purchase-requests',
          component: PurchaseRequestView,
          meta: { title: 'Purchase Requests' },
        },
        {
          path: 'inventory/purchase-orders',
          name: 'purchase-orders',
          component: PurchaseOrderView,
          meta: { title: 'Purchase Orders' },
        },
        {
          path: 'inventory/good-receipts',
          name: 'good-receipts',
          component: GoodReceiptView,
          meta: { title: 'Good Receipts' },
        },
        {
          path: 'inventory/goods-deliveries',
          name: 'goods-deliveries',
          component: GoodDeliveryView,
          meta: { title: 'Goods Deliveries' },
        },
        {
          path: 'inventory/stock',
          name: 'stock',
          component: StockView,
          meta: { title: 'Stock' },
        },
        {
          path: 'inventory/stock-adjustments',
          name: 'stock-adjustments',
          component: StockAdjustmentView,
          meta: { title: 'Stock Adjustments' },
        },
        {
          path: 'api-docs',
          name: 'api-docs',
          component: SwaggerView,
          meta: { title: 'API Documentation' },
        },
      ],
    },

    {
      path: '/:pathMatch(.*)*',
      redirect: '/dashboard',
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (auth.token && !auth.user) {
    try {
      await auth.fetchUser()
    } catch {
      // fetchUser sudah membersihkan token lokal jika gagal.
    }
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return '/login'
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return '/dashboard'
  }
})

export default router
