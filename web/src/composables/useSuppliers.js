import { onMounted, ref } from "vue";
import supplierService from "@/services/supplierService";

export function useSuppliers() {
  const suppliers = ref([]);
  const loading = ref(false);
  const submitting = ref(false);
  const deletingId = ref(null);
  const errorMessage = ref("");
  const modalError = ref("");
  const successMessage = ref("");
  const showModal = ref(false);
  const selectedSupplier = ref(null);

  async function fetchSuppliers() {
    loading.value = true;
    errorMessage.value = "";
    try {
      const response = await supplierService.getAll();
      suppliers.value = response.data.data || [];
    } catch (error) {
      errorMessage.value = error.response?.data?.message || "Failed to load suppliers.";
    } finally {
      loading.value = false;
    }
  }

  function openCreateModal() {
    successMessage.value = "";
    modalError.value = "";
    selectedSupplier.value = null;
    showModal.value = true;
  }

  async function openEditModal(id) {
    successMessage.value = "";
    modalError.value = "";
    try {
      const response = await supplierService.getById(id);
      selectedSupplier.value = response.data.data;
      showModal.value = true;
    } catch (error) {
      errorMessage.value = error.response?.data?.message || "Failed to load supplier details.";
    }
  }

  function closeModal() {
    if (submitting.value) return;
    showModal.value = false;
    selectedSupplier.value = null;
    modalError.value = "";
  }

  async function saveSupplier(payload) {
    submitting.value = true;
    modalError.value = "";
    try {
      if (selectedSupplier.value) {
        await supplierService.update(selectedSupplier.value.id, payload);
        successMessage.value = "Supplier updated successfully.";
      } else {
        await supplierService.create(payload);
        successMessage.value = "Supplier created successfully.";
      }
      showModal.value = false;
      selectedSupplier.value = null;
      await fetchSuppliers();
    } catch (error) {
      modalError.value = error.response?.data?.message || "Failed to save supplier.";
    } finally {
      submitting.value = false;
    }
  }

  async function deleteSupplier(supplier) {
    if (!window.confirm(`Are you sure you want to delete "${supplier.name}"?`)) return;

    deletingId.value = supplier.id;
    errorMessage.value = "";
    successMessage.value = "";
    try {
      await supplierService.delete(supplier.id);
      successMessage.value = "Supplier deleted successfully.";
      await fetchSuppliers();
    } catch (error) {
      errorMessage.value = error.response?.data?.message || "Failed to delete supplier.";
    } finally {
      deletingId.value = null;
    }
  }

  onMounted(fetchSuppliers);

  return {
    suppliers, loading, submitting, deletingId, errorMessage, modalError,
    successMessage, showModal, selectedSupplier, fetchSuppliers,
    openCreateModal, openEditModal, closeModal, saveSupplier, deleteSupplier,
  };
}
