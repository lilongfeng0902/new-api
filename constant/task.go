package constant

type TaskPlatform string

const (
	TaskPlatformSuno       TaskPlatform = "suno"
	TaskPlatformMidjourney              = "mj"
	TaskPlatformCqtai                   = "cqtai"
)

const (
	SunoActionMusic  = "MUSIC"
	SunoActionLyrics = "LYRICS"

	CqtaiActionMusic  = "MUSIC"
	CqtaiActionLyrics = "LYRICS"
	CqtaiActionFetch  = "FETCH"

	TaskActionGenerate          = "generate"
	TaskActionTextGenerate      = "textGenerate"
	TaskActionFirstTailGenerate = "firstTailGenerate"
	TaskActionReferenceGenerate = "referenceGenerate"
	TaskActionRemix             = "remixGenerate"
)

var SunoModel2Action = map[string]string{
	"suno_music":  SunoActionMusic,
	"suno_lyrics": SunoActionLyrics,
}

var CqtaiModel2Action = map[string]string{
	"suno_music":  CqtaiActionMusic,
	"suno_lyrics": CqtaiActionLyrics,
	"suno_fetch":  CqtaiActionFetch,
}
