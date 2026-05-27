package service

import (
	"encoding/json"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/galgame/model"
	"kun-galgame-api/pkg/userclient"
)

// decodeProviderNames safely turns the row's jsonb provider_name bytes into a
// string slice. A null/empty/invalid value yields an empty slice rather than
// nil so the JSON response stays `[]` (frontend-friendly).
func decodeProviderNames(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return []string{}
	}
	return out
}

// collectIDs extracts galgame IDs and user IDs from a list of resource rows.
func collectIDs(rows []model.GalgameResourceRow) (galgameIDs, userIDs []int) {
	galgameIDs = make([]int, 0, len(rows))
	userIDs = make([]int, 0, len(rows))
	for _, r := range rows {
		galgameIDs = append(galgameIDs, r.GalgameID)
		userIDs = append(userIDs, r.UserID)
	}
	return
}

// collectAggregate unions DISTINCT platform/language/type tuples into slices.
func collectAggregate(aggs []model.ResourceAggregate) (platforms, languages, types []string) {
	platforms, languages, types = []string{}, []string{}, []string{}
	for _, a := range aggs {
		if a.Platform != "" {
			platforms = appendUniqueStr(platforms, a.Platform)
		}
		if a.Language != "" {
			languages = appendUniqueStr(languages, a.Language)
		}
		if a.Type != "" {
			types = appendUniqueStr(types, a.Type)
		}
	}
	return
}

func appendUniqueStr(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}

// briefToName maps a wiki GalgameBrief to the four-language KunLanguage DTO.
func briefToName(b client.GalgameBrief) dto.KunLanguage {
	return dto.KunLanguage{
		EnUs: b.NameEnUs, JaJp: b.NameJaJp,
		ZhCn: b.NameZhCn, ZhTw: b.NameZhTw,
	}
}

// userBriefToDTO maps a userclient.User to the dto.UserBrief projection.
func userBriefToDTO(u userclient.User) dto.UserBrief {
	return dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar}
}

// rowToCard maps a resource row to the list-card DTO.
//
// `isLiked` is per-caller and must be looked up via the batch helper
// ResourceRepository.FindLikedSet at the service layer — the mapper
// itself has no DB access, so the caller threads the bool in. Was
// previously hardcoded `false`, which made every resource list (e.g.
// /galgame-resource and /galgame/:gid/resource/all) show the "已点赞"
// heart as off, even for users who had clicked it.
//
// `LinkDomain` is intentionally left empty on list cards — the row
// projection (`GalgameResourceRow`) doesn't carry the first link's
// domain because list endpoints don't join galgame_resource_link.
// FE falls back to `providerNames` for the display label; the rare
// case where providerNames is empty is acceptable absent a JOIN.
// Download-detail path (rowToDownloadDetail) DOES set this from
// `links[0]` because it already loads links.
func rowToCard(r model.GalgameResourceRow, u userclient.User, isLiked bool) dto.ResourceCard {
	return dto.ResourceCard{
		ID:            r.ID,
		View:          r.View,
		GalgameID:     r.GalgameID,
		User:          userBriefToDTO(u),
		Type:          r.Type,
		Language:      r.Language,
		Platform:      r.Platform,
		Size:          r.Size,
		Status:        r.Status,
		Download:      r.Download,
		LikeCount:     r.LikeCount,
		IsLiked:       isLiked,
		LinkDomain:    "",
		ProviderNames: decodeProviderNames(r.ProviderName),
		Note:          r.Note,
		Created:       r.Created,
		Edited:        r.Edited,
	}
}

// rowToDownloadDetail maps a resource row + links + liked flag + owner to the
// download-detail DTO.
func rowToDownloadDetail(
	r model.GalgameResourceRow,
	links []string,
	isLiked bool,
	owner userclient.User,
) dto.ResourceDownloadDetail {
	linkDomain := ""
	if len(links) > 0 {
		linkDomain = links[0]
	}
	return dto.ResourceDownloadDetail{
		ID:            r.ID,
		View:          r.View,
		GalgameID:     r.GalgameID,
		User:          userBriefToDTO(owner),
		Type:          r.Type,
		Language:      r.Language,
		Platform:      r.Platform,
		Size:          r.Size,
		Status:        r.Status,
		Download:      r.Download,
		LikeCount:     r.LikeCount,
		IsLiked:       isLiked,
		LinkDomain:    linkDomain,
		ProviderNames: decodeProviderNames(r.ProviderName),
		Link:          links,
		Code:          r.Code,
		Password:      r.Password,
		Note:          r.Note,
		Created:       r.Created,
		Edited:        r.Edited,
	}
}
