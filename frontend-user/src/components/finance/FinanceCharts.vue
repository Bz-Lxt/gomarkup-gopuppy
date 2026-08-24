<script setup lang="ts">
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, MarkAreaComponent, TooltipComponent } from 'echarts/components'
import { init, use, type ECharts } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ExpenseCategory, FinanceSummary } from '@/types/models'
import { EXPENSE_COLOR, EXPENSE_LABEL } from '@/utils/labels'

use([LineChart, BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent, MarkAreaComponent, CanvasRenderer])

const props = defineProps<{ summary: FinanceSummary }>()

const weightEl = ref<HTMLDivElement | null>(null)
const pieEl = ref<HTMLDivElement | null>(null)
const barEl = ref<HTMLDivElement | null>(null)
let weightChart: ECharts | null = null
let pieChart: ECharts | null = null
let barChart: ECharts | null = null

const cats: ExpenseCategory[] = ['FOOD', 'MEDICAL', 'TOY', 'GROOMING', 'INSURANCE', 'OTHER']

function axis() {
  return {
    axisLine: { lineStyle: { color: '#E4D2B8' } },
    axisLabel: { color: '#2A2118', fontFamily: 'Noto Serif SC' },
    splitLine: { lineStyle: { color: '#E4D2B8', type: 'dashed' } },
  }
}

function render() {
  const s = props.summary
  if (weightChart && weightEl.value) {
    const months = s.weight_series.map((w) => w.month)
    weightChart.setOption({
      color: ['#C45C26'],
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 16, top: 24, bottom: 28 },
      xAxis: { type: 'category', data: months, ...axis() },
      yAxis: { type: 'value', name: 'kg', ...axis() },
      series: [
        {
          type: 'line',
          smooth: true,
          symbolSize: 10,
          data: s.weight_series.map((w) => ({
            value: w.avg_kg,
            itemStyle: w.anomaly ? { color: '#B42318' } : { color: '#C45C26' },
          })),
          areaStyle: { color: 'rgba(196,92,38,0.08)' },
          markArea:
            s.weight_min != null && s.weight_max != null
              ? {
                  itemStyle: { color: 'rgba(61,107,79,0.08)' },
                  data: [[{ yAxis: s.weight_min }, { yAxis: s.weight_max }]],
                }
              : undefined,
        },
      ],
    })
  }
  if (pieChart) {
    pieChart.setOption({
      tooltip: { trigger: 'item' },
      series: [
        {
          type: 'pie',
          radius: ['42%', '68%'],
          label: { fontFamily: 'Noto Serif SC', color: '#2A2118' },
          data: s.pie.map((p) => ({
            name: EXPENSE_LABEL[p.category as ExpenseCategory] || p.category,
            value: p.cents / 100,
            itemStyle: { color: EXPENSE_COLOR[p.category as ExpenseCategory] || '#8A7A68' },
          })),
        },
      ],
    })
  }
  if (barChart) {
    const months = s.expense_series.map((e) => e.month)
    barChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { top: 0, textStyle: { fontFamily: 'Noto Serif SC', color: '#2A2118' } },
      grid: { left: 48, right: 12, top: 36, bottom: 28 },
      xAxis: { type: 'category', data: months, ...axis() },
      yAxis: { type: 'value', name: '元', ...axis() },
      series: cats.map((cat) => ({
        name: EXPENSE_LABEL[cat],
        type: 'bar',
        stack: 'exp',
        itemStyle: { color: EXPENSE_COLOR[cat] },
        data: s.expense_series.map((e) => (e.by_category?.[cat] || 0) / 100),
      })),
    })
  }
}

function resize() {
  weightChart?.resize()
  pieChart?.resize()
  barChart?.resize()
}

onMounted(async () => {
  await nextTick()
  if (weightEl.value) weightChart = init(weightEl.value)
  if (pieEl.value) pieChart = init(pieEl.value)
  if (barEl.value) barChart = init(barEl.value)
  render()
  window.addEventListener('resize', resize)
})

watch(() => props.summary, render, { deep: true })

onBeforeUnmount(() => {
  window.removeEventListener('resize', resize)
  weightChart?.dispose()
  pieChart?.dispose()
  barChart?.dispose()
})
</script>

<template>
  <div class="grid w-full gap-4 xl:grid-cols-2">
    <section class="rounded-page bg-card p-4 shadow-warm ring-1 ring-line xl:col-span-2">
      <h3 class="mb-2 font-display text-xl">体重曲线</h3>
      <div ref="weightEl" class="h-72 w-full" />
    </section>
    <section class="rounded-page bg-card p-4 shadow-warm ring-1 ring-line">
      <h3 class="mb-2 font-display text-xl">分类占比</h3>
      <div ref="pieEl" class="h-72 w-full" />
    </section>
    <section class="rounded-page bg-card p-4 shadow-warm ring-1 ring-line">
      <h3 class="mb-2 font-display text-xl">月度堆叠开销</h3>
      <div ref="barEl" class="h-72 w-full" />
    </section>
  </div>
</template>
