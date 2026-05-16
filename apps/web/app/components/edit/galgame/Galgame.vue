<script setup lang="ts">
// Submission form for the "new galgame" flow (POST /galgame/submit).
//
// No VNDB ID field: the wiki has already run `sync-vndb --full`, so
// every VNDB-listed work already exists there as a status=2 claimable
// draft. Anything reaching THIS form is by definition NOT in VNDB
// (original / doujin / indie), so a VNDB ID is meaningless here and
// would only collide with an existing draft (wiki 20004). VNDB works
// are handled by the publish wizard's claim flow, not submission.

import { languageItems } from '~/constants/edit'

const introductionLanguage = ref<Language>('zh-cn')

const { name } = storeToRefs(usePersistEditGalgameStore())
</script>

<template>
  <div class="contents">
    <ClientOnly>
      <KunCard
        :is-hoverable="false"
        :is-transparent="false"
        content-class="space-y-6"
      >
        <KunHeader
          name="提交 Galgame 申请"
          description="提交后将进入审核队列, 审核通过后才会被公开。如果该 Galgame 已收录于 VNDB 或已存在, 请通过「发布 Galgame」向导搜索并认领, 不要在此重复提交。您可以在「我的提交」页面查看进度、修改或撤回申请。"
        >
          <template #endContent>
            <KunLink target="_blank" to="/doc/galgame-publish-help">
              Galgame 发布系统帮助文档、教程、介绍
            </KunLink>
            <EditGalgameSite class="mt-2" />
          </template>
        </KunHeader>

        <KunInfo
          color="warning"
          title="此表单仅用于 VNDB 未收录的作品"
          description="Galgame Wiki 已全量同步 VNDB，VNDB 收录的作品都已是可认领草稿。请先在「发布 Galgame」向导按名称或 VNDB ID 搜索；搜索得到的就直接认领 / 发布资源，确实搜不到的原创 / 同人 / 独立作品再用本表单提交。"
        />

        <KunDivider>
          <span class="mx-2">必要信息</span>
        </KunDivider>

        <div class="space-y-2">
          <h2 class="text-xl">游戏名</h2>
          <p class="text-default-500 text-sm">
            游戏名要求至少写一种, 非常建议全部填写
            (如果游戏没有对应语言的官方译名可以不填写)
          </p>
          <div class="space-y-2">
            <KunInput placeholder="英语" v-model="name['en-us']" />
            <KunInput placeholder="日语" v-model="name['ja-jp']" />
            <KunInput placeholder="简体中文" v-model="name['zh-cn']" />
            <KunInput placeholder="繁体中文" v-model="name['zh-tw']" />
          </div>
        </div>

        <div class="space-y-2">
          <h2 class="text-xl">介绍</h2>
          <p class="text-default-500 text-sm">
            介绍要求至少写一种。可以参考官方页面、DLsite ストーリー、Bangumi
            或萌娘百科, 也可以抛弃官方介绍自己撰写。
          </p>
          <p class="text-default-500 text-sm">
            请不要在介绍部分放置任何 R18 图片, 我们的系统还没有开发完毕,
            之后会有专门处理图片的展示方式, 切记, 一定不要放置。R18
            的定义很严格, 过于露骨的图片都视为 R18。一般图片尽可能少的放置,
            能不放置图片就不要放置图片, 因为我们之后会专门做放图片的位置。
          </p>
          <KunTab
            :items="languageItems"
            v-model="introductionLanguage"
            variant="underlined"
            color="primary"
            size="sm"
          />
          <EditGalgameEditor :lang="introductionLanguage" type="publish" />
        </div>

        <EditGalgameBanner />

        <EditGalgameContentLimit type="create" />

        <EditGalgameMeta />

        <EditGalgamePrAlias type="create" />

        <KunInfo
          title="标签 / 会社 / 引擎"
          description="标签、会社、引擎等元数据由 Galgame Wiki 维护；审核通过后可在 Galgame 详情页提交 PR 补充。"
          color="info"
        />

        <EditGalgameFooter />
      </KunCard>
    </ClientOnly>
  </div>
</template>
