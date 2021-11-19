package helpers

const (
	XrfKeySize = 16
)

var (
	XrfKeyChars = []rune("abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ0123456789")
)

func GenerateXrfKey(rnd Randomizer) string {
	key := ""
	for i := 0; i < XrfKeySize; i++ {
		key += string(rnd.RandRune(XrfKeyChars))
	}
	return key
}
