package browser

import (
	"fmt"
	"time"
)

type BrowserPreferences struct {
	ShowHomeButton                    bool     `json:"show_home_button"`
	CheckDefaultBrowser               bool     `json:"check_default_browser"`
	DefaultBrowserInfobarLastDeclined string   `json:"default_browser_infobar_last_declined"`
	EnabledLabsExperiments            []string `json:"enabled_labs_experiments,omitempty"`
}

type SessionPreferences struct {
	RestoreOnStartup int      `json:"restore_on_startup"`
	StartupUrls      []string `json:"startup_urls,omitempty"`
}

type BookmarkBarPreferences struct {
	ShowOnAllTabs bool `json:"show_on_all_tabs"`
}

type SyncPromoPreferences struct {
	ShowOnFirstRunAllowed bool `json:"show_on_first_run_allowed"`
}

type DistributionPreferences struct {
	DoNotCreateAnyShortcuts              bool   `json:"do_not_create_any_shortcuts"`
	DoNotCreateDesktopShortcut           bool   `json:"do_not_create_desktop_shortcut"`
	DoNotCreateQuickLaunchShortcut       bool   `json:"do_not_create_quick_launch_shortcut"`
	DoNotCreateTaskbarShortcut           bool   `json:"do_not_create_taskbar_shortcut"`
	DoNotLaunchChrome                    bool   `json:"do_not_launch_chrome"`
	DoNotRegisterForUpdateLaunch         bool   `json:"do_not_register_for_update_launch"`
	ImportBookmarks                      bool   `json:"import_bookmarks"`
	ImportBookmarksFromFile              string `json:"import_bookmarks_from_file,omitempty"`
	ImportHistory                        bool   `json:"import_history"`
	ImportHomePage                       bool   `json:"import_home_page"`
	ImportSearchEngine                   bool   `json:"import_search_engine"`
	MakeChromeDefault                    bool   `json:"make_chrome_default"`
	MakeChromeDefaultForUser             bool   `json:"make_chrome_default_for_user"`
	PingDelay                            int    `json:"ping_delay,omitempty"`
	RequireEula                          bool   `json:"require_eula"`
	SuppressFirstRunBubble               bool   `json:"suppress_first_run_bubble"`
	SuppressFirstRunDefaultBrowserPrompt bool   `json:"suppress_first_run_default_browser_prompt"`
	SystemLevel                          bool   `json:"system_level"`
	VerboseLogging                       bool   `json:"verbose_logging"`
}

type Preferences struct {
	Homepage             string                   `json:"homepage"`
	HomepageIsNewTabPage bool                     `json:"homepage_is_newtabpage"`
	Browser              *BrowserPreferences      `json:"browser,omitempty"`
	Session              *SessionPreferences      `json:"session,omitempty"`
	BookmarkBar          *BookmarkBarPreferences  `json:"bookmark_bar,omitempty"`
	SyncPromo            *SyncPromoPreferences    `json:"sync_promo,omitempty"`
	Distribution         *DistributionPreferences `json:"distribution,omitempty"`
	FirstRunTabs         []string                 `json:"first_run_tabs,omitempty"`
}

func GetDefaultPreferences() *Preferences {
	return &Preferences{
		Homepage:             `about:blank`,
		HomepageIsNewTabPage: true,
		Browser: &BrowserPreferences{
			ShowHomeButton:                    false,
			CheckDefaultBrowser:               false,
			DefaultBrowserInfobarLastDeclined: fmt.Sprintf("%v", time.Now().UnixNano()),
			EnabledLabsExperiments: []string{
				`overscroll-history-navigation@1`,
			},
		},
		Session: &SessionPreferences{
			RestoreOnStartup: 1,
			StartupUrls:      []string{`about:blank`},
		},
		BookmarkBar: &BookmarkBarPreferences{
			ShowOnAllTabs: false,
		},
		SyncPromo: &SyncPromoPreferences{
			ShowOnFirstRunAllowed: false,
		},
		Distribution: &DistributionPreferences{
			DoNotCreateAnyShortcuts:              true,
			DoNotCreateDesktopShortcut:           true,
			DoNotCreateQuickLaunchShortcut:       true,
			DoNotCreateTaskbarShortcut:           true,
			DoNotLaunchChrome:                    true,
			DoNotRegisterForUpdateLaunch:         true,
			ImportBookmarks:                      false,
			ImportHistory:                        false,
			ImportHomePage:                       true,
			ImportSearchEngine:                   false,
			MakeChromeDefault:                    false,
			MakeChromeDefaultForUser:             false,
			PingDelay:                            60,
			RequireEula:                          false,
			SuppressFirstRunBubble:               true,
			SuppressFirstRunDefaultBrowserPrompt: true,
			SystemLevel:                          false,
			VerboseLogging:                       false,
		},
	}
}
