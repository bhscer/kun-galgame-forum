<script setup lang="ts">
const props = defineProps<{
  galgame: GalgameDetail
}>()

provide<GalgameDetail>('galgame', props.galgame)

const ratings = ref([...props.galgame.ratings])
const sortedRatings = computed(() => {
  return [...ratings.value].sort(
    (a, b) => b.short_summary.length - a.short_summary.length
  )
})

const handleRatingCreated = (newRating: GalgameRatingCardOnGalgamePage) => {
  ratings.value.unshift(newRating)
}
</script>

<template>
  <div class="space-y-3">
    <GalgameHeader
      :galgame="galgame"
      @on-rating-created="handleRatingCreated"
    />

    <div
      v-if="sortedRatings.length && sortedRatings.length >= 3"
      class="grid grid-cols-1 gap-3"
    >
      <GalgameRatingRadarCard :ratings="sortedRatings" />
    </div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
      <div class="md:col-span-2">
        <KunCard
          :is-hoverable="false"
          :is-transparent="false"
          content-class="space-y-12 relative"
        >
          <div class="space-y-3">
            <GalgameIntroduction :introduction="galgame.introduction" />

            <div
              v-if="sortedRatings.length && sortedRatings.length < 3"
              class="space-y-1"
            >
              <GalgameRatingRow
                v-for="rating in sortedRatings"
                :key="rating.id"
                :rating="rating"
              />
            </div>

            <GalgameLink />
          </div>

          <GalgameGallery :screenshots="galgame.screenshots" />

          <GalgameResource />

          <GalgamePatchContainer :vndb-id="galgame.vndbId" />

          <div v-if="galgame.series" class="space-y-3">
            <KunHeader
              name="Galgame 系列"
              description="Galgame 全系列所有 Galgame 作品。例如美少女万华镜 1, 2, 3, 4, 5, 雪女, 外传 就是一个 Galgame 系列"
              scale="h3"
            />
            <GalgameSeriesCard :series="galgame.series" />
          </div>

          <GalgameCommentContainer
            :user-data="galgame.contributor"
            :target-user="galgame.user"
          />
        </KunCard>
      </div>

      <div class="space-y-3 md:col-span-1">
        <GalgameTag :tags="galgame.tag" />

        <GalgameInfo
          :official="galgame.official"
          :engine="galgame.engine"
          :age-limit="galgame.ageLimit"
          :original-language="galgame.originalLanguage"
          :release-date="galgame.releaseDate"
          :release-date-tba="galgame.releaseDateTBA"
        />

        <KunCard
          content-class="space-y-3"
          :is-hoverable="false"
          :is-transparent="false"
        >
          <KunHeader
            name="贡献者"
            description="本游戏项目的贡献者, 计 Galgame 资源发布贡献"
            scale="h3"
          />

          <div
            class="text-default-500 flex cursor-default flex-wrap items-center gap-2"
          >
            <KunUser :user="galgame.user" />
            <span class="text-sm">
              <KunTime :time="galgame.created" type="date" show-year /> 创建本游戏
            </span>
          </div>

          <GalgameContributorContainer />
        </KunCard>

        <div class="text-default-500 flex items-center justify-center text-sm">
          部分页面数据由
          <KunLink size="sm" target="_blank" to="https://vndb.org">
            VNDB
          </KunLink>
          提供
        </div>
      </div>
    </div>
  </div>
</template>
