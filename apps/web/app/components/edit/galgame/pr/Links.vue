<script setup lang="ts">
// Related-links editor for the PR form. Wiki PR `links` is replace-all:
// the rows here MUST be pre-hydrated with the galgame's CURRENT links
// (see components/galgame/Rewrite.vue, GET /galgame/:gid/link/all) or a
// submit silently wipes every existing link. Empty rows are dropped at
// submit (Footer.vue filters blank name/link).
const { galgamePR } = storeToRefs(useTempGalgamePRStore())

const links = computed<{ name: string; link: string }[]>({
  get: () => galgamePR.value[0]?.links ?? [],
  set: (v) => {
    if (galgamePR.value[0]) galgamePR.value[0].links = v
  }
})

const addRow = () => {
  links.value = [...links.value, { name: '', link: '' }]
}

const removeRow = (index: number) => {
  links.value = links.value.filter((_, i) => i !== index)
}
</script>

<template>
  <div class="space-y-2">
    <h2 class="text-xl">相关链接</h2>
    <p class="text-default-500 text-sm">
      官网 / 商店 / 资料页等。提交后将整组替换该 Galgame 的相关链接,
      留空的行会被忽略。
    </p>

    <div
      v-for="(row, index) in links"
      :key="index"
      class="flex flex-col gap-2 sm:flex-row sm:items-center"
    >
      <KunInput
        v-model="row.name"
        placeholder="名称 (如: 官方网站)"
        class="sm:w-1/3"
      />
      <KunInput
        v-model="row.link"
        placeholder="https://..."
        class="flex-1"
      />
      <KunButton
        :is-icon-only="true"
        variant="flat"
        color="danger"
        @click="removeRow(index)"
      >
        <KunIcon name="lucide:trash-2" />
      </KunButton>
    </div>

    <KunButton variant="flat" @click="addRow">
      <KunIcon name="lucide:plus" />
      添加链接
    </KunButton>
  </div>
</template>
