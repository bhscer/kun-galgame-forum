<script setup lang="ts">
// PR edit form for a PUBLISHED galgame (POST /galgame/:gid/prs). Single
// page, sectioned with KunDivider — deliberately NOT a stepper: editing
// is "tweak a field or two on fully-known data", and the wiki's
// replace-all semantics mean every relation section must stay visible &
// pre-filled at once so the user can see "this is the current set".
// Aligns with Draft.vue's KunCard + KunDivider look.
//
// Single root is <div class="contents"> (page renders us bare); the
// galgamePR[0] guard handles the persist:false edge (browser back into
// an emptied temp store) — the normal entry (components/galgame/
// Rewrite.vue) always hydrates it first, and ?type=pr arms the
// `prevent` middleware against hard refreshes.
import type { KunSelectOption } from '~/components/kun/select/type'
import { languageItems } from '~/constants/edit'

const { galgamePR } = storeToRefs(useTempGalgamePRStore())
const pr = computed(() => galgamePR.value[0])

const introductionLanguage = ref<Language>('zh-cn')

// Typed as KunSelectOption[] (not `as const`): KunSelect.options is a
// mutable KunSelectOption[], a readonly tuple from `as const` is not
// assignable. Matches the convention in app/constants/*.ts.
const ageLimitOptions: KunSelectOption[] = [
  { value: 'all', label: '全年龄' },
  { value: 'r18', label: 'R18' }
]

const originalLanguageOptions: KunSelectOption[] = [
  { value: 'ja-jp', label: '日语' },
  { value: 'en-us', label: '英语' },
  { value: 'zh-cn', label: '简体中文' },
  { value: 'zh-tw', label: '繁体中文' },
  { value: 'others', label: '其它' }
]
</script>

<template>
  <div class="contents">
    <ClientOnly>
      <KunCard
        :is-hoverable="false"
        :is-pressable="false"
        :is-transparent="false"
        content-class="space-y-6"
      >
        <EditGalgamePrHelp />

        <KunInfo
          v-if="!pr"
          color="warning"
          title="无法编辑"
          description="请从 Galgame 详情页点击编辑按钮进入此页面。"
        />

        <template v-else>
          <KunDivider>
            <span class="mx-2">基本信息</span>
          </KunDivider>

          <EditGalgamePrTitle />

          <KunDivider>
            <span class="mx-2">游戏介绍</span>
          </KunDivider>

          <div class="space-y-2">
            <KunTab
              :items="languageItems"
              v-model="introductionLanguage"
              variant="underlined"
              color="primary"
              size="sm"
            />
            <EditGalgameEditor :lang="introductionLanguage" type="rewrite" />
          </div>

          <KunDivider>
            <span class="mx-2">别名</span>
          </KunDivider>

          <EditGalgamePrAlias type="rewrite" />

          <KunDivider>
            <span class="mx-2">标签 / 会社 / 引擎</span>
          </KunDivider>

          <EditGalgamePrTagSelector />
          <EditGalgamePrOfficialSelector />
          <EditGalgamePrEngineSelector />

          <KunDivider>
            <span class="mx-2">相关链接</span>
          </KunDivider>

          <EditGalgamePrLinks />

          <KunDivider>
            <span class="mx-2">分级与元数据</span>
          </KunDivider>

          <EditGalgameContentLimit type="rewrite" />

          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <KunSelect
              v-model="pr!.ageLimit"
              label="年龄分级"
              :options="ageLimitOptions"
            />
            <KunSelect
              v-model="pr!.originalLanguage"
              label="原始语言"
              :options="originalLanguageOptions"
            />
          </div>

          <KunDivider>
            <span class="mx-2">高级 (可选)</span>
          </KunDivider>

          <div class="space-y-2">
            <h2 class="text-xl">VNDB 编号</h2>
            <p class="text-default-500 text-sm">
              该作品在 VNDB 的编号, 无对应条目 (如原创 / 同人作品) 可留空。
            </p>
            <KunInput
              v-model="pr!.vndbId"
              placeholder="例如: v19658 (可留空)"
            />
          </div>

          <div class="space-y-2">
            <h2 class="text-xl">PR 说明</h2>
            <p class="text-default-500 text-sm">
              简述本次修改了什么、依据来源, 便于审核者快速判断是否合并。
            </p>
            <KunTextarea
              v-model="pr!.note"
              placeholder="例如: 补充标签与开发商, 修正简体中文简介"
            />
          </div>

          <KunDivider>
            <span class="mx-2">封面</span>
          </KunDivider>

          <EditGalgameBanner />

          <EditGalgamePrFooter />
        </template>
      </KunCard>
    </ClientOnly>
  </div>
</template>
