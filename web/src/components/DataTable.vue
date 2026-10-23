<template>
  <div class="space-y-4">
    <div class="rounded-md border">
      <table class="w-full caption-bottom text-sm">
        <thead class="[&_tr]:border-b">
          <tr
            v-for="headerGroup in table.getHeaderGroups()"
            :key="headerGroup.id"
            class="border-b transition-colors hover:bg-muted/50"
          >
            <th
              v-for="header in headerGroup.headers"
              :key="header.id"
              class="h-12 px-4 text-left align-middle font-medium text-muted-foreground"
            >
              <div
                v-if="header.column.getCanSort()"
                class="flex items-center gap-2 cursor-pointer select-none"
                @click="header.column.toggleSorting()"
              >
                <component
                  :is="
                    header.column.getIsSorted() === false
                      ? ChevronsUpDownIcon
                      : header.column.getIsSorted() === 'asc'
                        ? ChevronUpIcon
                        : ChevronDownIcon
                  "
                  v-if="header.column.getCanSort()"
                  class="h-4 w-4"
                  :class="
                    header.column.getIsSorted() ? 'opacity-100' : 'opacity-50'
                  "
                />
                <span>{{
                  typeof header.column.columnDef.header === "function"
                    ? ""
                    : header.column.columnDef.header
                }}</span>
              </div>
              <span v-else>{{
                typeof header.column.columnDef.header === "function"
                  ? ""
                  : header.column.columnDef.header
              }}</span>
            </th>
          </tr>
        </thead>
        <tbody class="[&_tr:last-child]:border-0">
          <tr
            v-for="row in table.getRowModel().rows"
            :key="row.id"
            class="border-b transition-colors hover:bg-muted/50"
          >
            <td
              v-for="cell in row.getVisibleCells()"
              :key="cell.id"
              class="p-4 align-middle"
            >
              <component
                :is="cell.column.columnDef.cell"
                v-if="cell.column.columnDef.cell"
                v-bind="cell.getContext()"
              />
              <span v-else>{{ cell.getValue() }}</span>
            </td>
          </tr>
          <tr v-if="table.getRowModel().rows.length === 0">
            <td :colspan="columns.length" class="p-8 text-center text-muted-foreground">
              No data available
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex items-center justify-between px-2">
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">
          Page {{ pageIndex + 1 }} of {{ pageCount }}
        </span>
        <select
          :value="pageSize"
          class="h-8 w-[70px] rounded-md border border-input bg-background px-2 text-sm"
          @change="
            table.setPageSize(
              Number(($event.target as HTMLSelectElement).value),
            )
          "
        >
          <option :value="10">10</option>
          <option :value="20">20</option>
          <option :value="50">50</option>
        </select>
      </div>

      <div class="flex items-center gap-1">
        <Button
          variant="outline"
          size="sm"
          :disabled="!table.getCanPreviousPage()"
          @click="table.setPageIndex(0)"
        >
          <ChevronsLeftIcon class="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="!table.getCanPreviousPage()"
          @click="table.previousPage()"
        >
          <ChevronLeftIcon class="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="!table.getCanNextPage()"
          @click="table.nextPage()"
        >
          <ChevronRightIcon class="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="!table.getCanNextPage()"
          @click="table.setPageIndex(pageCount - 1)"
        >
          <ChevronsRightIcon class="h-4 w-4" />
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import {
  useTable,
  tableFeatures,
  columnVisibilityFeature,
  rowPaginationFeature,
  rowSortingFeature,
  createPaginatedRowModel,
  createSortedRowModel,
  type SortingState,
  type ColumnDef,
  type Updater,
  type PaginationState,
} from "@tanstack/vue-table"
import {
  ChevronDownIcon,
  ChevronUpIcon,
  ChevronsUpDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronsLeftIcon,
  ChevronsRightIcon,
} from "lucide-vue-next"
import { Button } from "@/components/ui/button"

const props = defineProps<{
  columns: ColumnDef<any, any>[]
  data: any[]
  pageSize?: number
}>()

const sorting = ref<SortingState>([])
const pagination = ref<PaginationState>({
  pageIndex: 0,
  pageSize: props.pageSize || 10,
})

const features = tableFeatures({
  columnVisibilityFeature,
  rowPaginationFeature,
  rowSortingFeature,
  sortedRowModel: createSortedRowModel(),
  paginatedRowModel: createPaginatedRowModel(),
})

const columns = computed(() => props.columns)
const data = computed(() => props.data)
const state = computed(() => ({
  sorting: sorting.value,
  pagination: pagination.value,
}))

const table = useTable({
  features,
  columns,
  data,
  state,
  onSortingChange: (updater: Updater<SortingState>) => {
    sorting.value =
      typeof updater === "function" ? updater(sorting.value) : updater
  },
  onPaginationChange: (updater: Updater<PaginationState>) => {
    pagination.value =
      typeof updater === "function" ? updater(pagination.value) : updater
  },
})

const pageCount = computed(() => table.getPageCount())
const pageIndex = computed(() => pagination.value.pageIndex)
const pageSize = computed(() => pagination.value.pageSize)
</script>
