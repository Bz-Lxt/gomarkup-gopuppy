<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { mediaApi } from '@/api/endpoints'
import { apiErrorMessage } from '@/api/http'
import AppShell from '@/components/layout/AppShell.vue'
import Lightbox from '@/components/media/Lightbox.vue'
import PhotoMasonry from '@/components/media/PhotoMasonry.vue'
import AppButton from '@/components/ui/AppButton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SkeletonCard from '@/components/ui/SkeletonCard.vue'
import { downloadMedia } from '@/composables/useMediaUrl'
import { useFamilyStore } from '@/stores/family'
import { useToastStore } from '@/stores/toast'
import type { MediaFile, MediaKind } from '@/types/models'
import { formatDateTime } from '@/utils/datetime'
import { MEDIA_LABEL, bytes } from '@/utils/labels'

const family = useFamilyStore()
const toast = useToastStore()
const kind = ref<MediaKind>('PHOTO')
const items = ref<MediaFile[]>([])
const loading = ref(false)
const lightbox = ref(-1)
const uploading = ref(false)

const photos = computed(() => items.value.filter((m) => m.mime.startsWith('image/')))
const docs = computed(() => items.value.filter((m) => !m.mime.startsWith('image/')))

async function load() {
  const pet = family.currentPet
  if (!pet) {
    items.value = []
    return
  }
  loading.value = true
  try {
    items.value = await mediaApi.list(pet.id, kind.value)
  } catch (e) {
    toast.error(apiErrorMessage(e))
  } finally {
    loading.value = false
  }
}

function whenReady() {
  if (family.bootstrapped) void load()
  else {
    const stop = watch(
      () => family.bootstrapped,
      (ok) => {
        if (ok) {
          void load()
          stop()
        }
      },
    )
  }
}

watch(
  () => [family.currentPetId, kind.value],
  () => {
    if (family.bootstrapped) void load()
  },
)
onMounted(whenReady)

async function onFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !family.currentPet) return
  if (file.size > 20 * 1024 * 1024) {
    toast.error('单文件不超过 20MB')
    return
  }
  uploading.value = true
  try {
    await mediaApi.upload(family.currentPet.id, file, kind.value)
    toast.success('已收入相册')
    await load()
  } catch (e) {
    toast.error(apiErrorMessage(e))
  } finally {
    uploading.value = false
  }
}

async function saveDoc(m: MediaFile) {
  try {
    await downloadMedia(m.id, m.filename)
    toast.success('开始下载')
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}

async function previewPdf(m: MediaFile) {
  try {
    const res = await mediaApi.file(m.id)
    const url = URL.createObjectURL(res.data)
    window.open(url, '_blank')
  } catch (e) {
    toast.error(apiErrorMessage(e))
  }
}
</script>

<template>
  <AppShell>
    <div class="mb-5 flex flex-wrap items-end justify-between gap-3">
      <div>
        <p class="text-xs tracking-[0.25em] text-ink/40">ALBUM</p>
        <h1 class="font-display text-4xl">{{ family.currentPet?.name || '相册' }} 的云端相册</h1>
      </div>
      <div class="flex flex-wrap gap-2">
        <select
          v-if="family.pets.length"
          class="rounded-2xl bg-card px-3 py-2 text-sm ring-1 ring-line"
          :value="family.currentPetId"
          @change="family.setPet(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="p in family.pets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <label class="stamp-press inline-flex cursor-pointer items-center rounded-2xl bg-clay px-4 py-2.5 text-sm text-white shadow-stamp">
          {{ uploading ? '上传中…' : '上传文件' }}
          <input type="file" class="hidden" :disabled="!family.canWrite || uploading" @change="onFile" />
        </label>
      </div>
    </div>

    <div class="mb-5 flex gap-2">
      <button
        v-for="k in (['PHOTO', 'MEDICAL_RECORD', 'REPORT_PDF'] as MediaKind[])"
        :key="k"
        type="button"
        class="rounded-full px-3 py-1 text-sm ring-1"
        :class="kind === k ? 'bg-clay text-white ring-clay' : 'bg-card ring-line'"
        @click="kind = k"
      >
        {{ MEDIA_LABEL[k] }}
      </button>
    </div>

    <SkeletonCard v-if="loading" />
    <EmptyState v-else-if="!items.length" title="相册还是空白页" hint="JPEG / PNG / WEBP / PDF，单文件不超过 20MB。" />
    <template v-else>
      <PhotoMasonry v-if="photos.length" :items="photos" :key="kind + photos.map((p) => p.id).join()" @open="lightbox = $event" />
      <div v-if="docs.length" class="mt-6 space-y-2">
        <article
          v-for="m in docs"
          :key="m.id"
          class="flex flex-wrap items-center justify-between gap-3 rounded-page bg-card px-4 py-3 shadow-warm ring-1 ring-line"
        >
          <div>
            <p class="font-medium">{{ m.filename }}</p>
            <p class="text-xs text-ink/45">{{ bytes(m.size_bytes) }} · {{ formatDateTime(m.created_at) }}</p>
          </div>
          <div class="flex gap-2">
            <AppButton variant="ghost" @click="previewPdf(m)">预览</AppButton>
            <AppButton @click="saveDoc(m)">下载 PDF</AppButton>
          </div>
        </article>
      </div>
    </template>

    <Lightbox
      v-if="lightbox >= 0 && photos[lightbox]"
      :items="photos"
      :index="lightbox"
      @close="lightbox = -1"
      @index="lightbox = $event"
    />
  </AppShell>
</template>
