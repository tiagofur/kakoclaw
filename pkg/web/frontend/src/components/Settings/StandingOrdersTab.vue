<template>
  <div class="space-y-5 animate-fade-in-up">
    <!-- Standing Orders Panel -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:border-makoclaw-accent/30 hover:shadow-lg">
      <div class="flex items-center gap-3 mb-5">
        <div class="p-2 rounded-xl bg-makoclaw-accent/10 text-makoclaw-accent">
          <ClipboardDocumentListIcon class="w-5 h-5" />
        </div>
        <div>
          <h3 class="text-[10px] font-bold uppercase tracking-[0.15em] text-makoclaw-accent leading-none mb-1">
            Persistent Instructions
          </h3>
          <p class="text-lg font-black text-makoclaw-text uppercase italic tracking-tight">
            Standing Orders
          </p>
        </div>
      </div>

      <!-- Loading -->
      <div
        v-if="loading"
        class="flex items-center justify-center py-12"
      >
        <div class="w-10 h-10 border-4 border-makoclaw-accent/20 border-t-makoclaw-accent rounded-full animate-spin" />
      </div>

      <div
        v-else
        class="space-y-4"
      >
        <!-- Error / Success messages -->
        <div
          v-if="errorMessage"
          class="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-xs font-medium"
        >
          {{ errorMessage }}
        </div>
        <div
          v-if="successMessage"
          class="p-3 bg-makoclaw-success/10 border border-makoclaw-success/20 rounded-xl text-makoclaw-success text-xs font-medium"
        >
          {{ successMessage }}
        </div>

        <!-- Orders List -->
        <div
          v-if="orders.length === 0"
          class="text-center py-8 text-makoclaw-text-secondary text-sm"
        >
          No standing orders yet. Add one below.
        </div>

        <div
          v-for="order in orders"
          :key="order.id"
          class="flex items-start gap-3 px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30 group"
        >
          <!-- Active indicator -->
          <div
            class="mt-1 w-2 h-2 rounded-full flex-none"
            :class="order.active ? 'bg-makoclaw-success' : 'bg-makoclaw-text-secondary/30'"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm text-makoclaw-text font-medium break-words">
              {{ order.content }}
            </p>
            <p
              v-if="order.label"
              class="text-[10px] text-makoclaw-text-secondary/60 mt-0.5 font-medium uppercase tracking-wide"
            >
              {{ order.label }}
            </p>
          </div>
          <button
            class="opacity-0 group-hover:opacity-100 transition-opacity p-1.5 rounded-lg hover:bg-red-500/10 text-makoclaw-text-secondary hover:text-red-400"
            title="Delete"
            @click="confirmDelete(order)"
          >
            <TrashIcon class="w-4 h-4" />
          </button>
        </div>

        <!-- Add Form -->
        <div class="pt-2 border-t border-makoclaw-border/30 space-y-3">
          <p class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/60">
            New Standing Order
          </p>
          <div>
            <textarea
              v-model="newContent"
              rows="3"
              placeholder="Always respond in bullet points..."
              maxlength="200"
              class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-3 text-sm text-makoclaw-text focus:border-makoclaw-accent/50 focus:outline-none transition-all resize-none"
            />
            <p class="text-[10px] text-makoclaw-text-secondary/40 mt-0.5 text-right">
              {{ newContent.length }}/200
            </p>
          </div>
          <input
            v-model="newLabel"
            type="text"
            placeholder="Label (optional)"
            class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm text-makoclaw-text focus:border-makoclaw-accent/50 focus:outline-none transition-all"
          >
          <div
            v-if="addError"
            class="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-xs font-medium"
          >
            {{ addError }}
          </div>
          <button
            :disabled="saving || !newContent.trim()"
            class="px-4 py-2 rounded-xl bg-makoclaw-accent/10 text-makoclaw-accent border border-makoclaw-accent/20 text-xs font-medium hover:bg-makoclaw-accent hover:text-white transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="addOrder"
          >
            {{ saving ? 'Adding...' : 'Add Order' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showDeleteModal"
          class="fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
          @click.self="showDeleteModal = false"
        >
          <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-sm animate-scaleIn p-5 space-y-4">
            <h3 class="text-base font-bold text-makoclaw-text">
              Delete Standing Order
            </h3>
            <p class="text-sm text-makoclaw-text-secondary">
              Are you sure you want to delete this order?
            </p>
            <p
              v-if="orderToDelete"
              class="text-xs font-medium text-makoclaw-text bg-makoclaw-bg/40 rounded-lg px-3 py-2 border border-makoclaw-border/30"
            >
              "{{ orderToDelete.content }}"
            </p>
            <div class="flex gap-3 justify-end">
              <button
                class="px-4 py-2 rounded-xl text-xs font-medium text-makoclaw-text-secondary hover:bg-makoclaw-surface-hover transition-all"
                @click="showDeleteModal = false"
              >
                Cancel
              </button>
              <button
                :disabled="deleting"
                class="px-4 py-2 rounded-xl bg-red-500/10 text-red-400 border border-red-500/20 text-xs font-medium hover:bg-red-500 hover:text-white transition-all active:scale-95 disabled:opacity-50"
                @click="deleteOrder"
              >
                {{ deleting ? 'Deleting...' : 'Delete' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ClipboardDocumentListIcon, TrashIcon } from '@heroicons/vue/24/outline'
import standingOrdersService from '../../services/standingOrders'

const loading = ref(true)
const saving = ref(false)
const deleting = ref(false)
const orders = ref([])
const newContent = ref('')
const newLabel = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const addError = ref('')
const showDeleteModal = ref(false)
const orderToDelete = ref(null)

const fetchOrders = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await standingOrdersService.listOrders()
    orders.value = data.standing_orders || []
  } catch {
    errorMessage.value = 'Failed to load standing orders.'
  } finally {
    loading.value = false
  }
}

const addOrder = async () => {
  addError.value = ''
  if (!newContent.value.trim()) return
  saving.value = true
  try {
    await standingOrdersService.createOrder(newContent.value.trim(), newLabel.value.trim())
    newContent.value = ''
    newLabel.value = ''
    successMessage.value = 'Standing order added.'
    setTimeout(() => { successMessage.value = '' }, 3000)
    await fetchOrders()
  } catch (err) {
    const msg = err?.response?.data || err?.message || 'Failed to add standing order.'
    addError.value = typeof msg === 'string' ? msg.trim() : 'Failed to add standing order.'
  } finally {
    saving.value = false
  }
}

const confirmDelete = (order) => {
  orderToDelete.value = order
  showDeleteModal.value = true
}

const deleteOrder = async () => {
  if (!orderToDelete.value) return
  deleting.value = true
  try {
    await standingOrdersService.deleteOrder(orderToDelete.value.id)
    showDeleteModal.value = false
    orderToDelete.value = null
    successMessage.value = 'Standing order deleted.'
    setTimeout(() => { successMessage.value = '' }, 3000)
    await fetchOrders()
  } catch {
    errorMessage.value = 'Failed to delete standing order.'
    showDeleteModal.value = false
  } finally {
    deleting.value = false
  }
}

onMounted(fetchOrders)
</script>
