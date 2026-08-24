<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'clay' | 'moss' | 'ghost' | 'danger'
    disabled?: boolean
    type?: 'button' | 'submit'
  }>(),
  { variant: 'clay', type: 'button' },
)

const cls: Record<string, string> = {
  clay: 'bg-clay text-white shadow-stamp hover:brightness-105 disabled:bg-clay/40',
  moss: 'bg-moss text-white shadow-stamp hover:brightness-105 disabled:bg-moss/40',
  ghost: 'bg-transparent text-ink ring-1 ring-line hover:bg-paper disabled:opacity-40',
  danger: 'bg-rose text-white shadow-stamp hover:brightness-105 disabled:bg-rose/40',
}
</script>

<template>
  <button
    :type="type"
    :disabled="disabled"
    class="stamp-press inline-flex items-center justify-center rounded-2xl px-4 py-2.5 text-sm tracking-wide disabled:cursor-not-allowed disabled:shadow-none"
    :class="cls[variant]"
  >
    <slot />
  </button>
</template>
