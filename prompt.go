package issh

// Prompt represent ssh prompt pettern used in check whether command is finished.
// Example:
//   Terminal Prompt "pi@raspberrypi:~ $ "
//     SufixPattern='$' (rune to byte)
//     SufixPosition=2 ($ + space)
//
// When you use multi prompt such as Root user (often '#'), you must give all prompt pattern before Run
type Prompt struct {
	SufixPattern  byte
	SufixPosition int
}

var (
	// DefaultPrompt is prompt pettern like "pi@raspberrypi:~ $ "
	DefaultPrompt = Prompt{
		SufixPattern:  '$',
		SufixPosition: 2,
	}
	// DefaultRootPrompt is prompt pettern like "# "
	DefaultRootPrompt = Prompt{
		SufixPattern:  '#',
		SufixPosition: 2,
	}
)
