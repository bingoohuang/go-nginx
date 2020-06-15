package directive

import "github.com/pkg/errors"

// ModifierPriority defines the priority of the modifier.
// https://end0tknr.wordpress.com/2015/12/22/location-match-priority-in-nginx/.
// https://artfulrobot.uk/blog/untangling-nginx-location-block-matching-algorithm.
// https://blog.csdn.net/qq_15766181/article/details/72829672.
// priority	| prefix       	                          | example.
// 1        | = (exactly)	                          | location = /path.
// 2        | ^~ (forward match)	                  | location = /image.
// 3        | ~ or ~* (regular & case-(in)sensitive)  | location ~ /image/.
// 4        | NONE (forward match)                    | location /image.
type ModifierPriority int

// https://stackoverflow.com/questions/5238377/nginx-location-priority
const (
	// ModifierExactly like location = /path matches the query / only.
	ModifierExactly ModifierPriority = iota + 1
	// ModifierForward like location ^~ /path.
	// matches any query beginning with /path/ and halts searching,
	// so regular expressions will not be checked.
	ModifierForward
	// ModifierRegular like location ~ \.(gif|jpg|jpeg)$
	// or like location ~* .(jpg|png|bmp).
	ModifierRegular
	// ModifierNone means none modifier for the location.
	ModifierNone
)

// Modifier is the location modifier.
type Modifier string

// Priority returns the priority of the location matching.
func (m Modifier) Priority() ModifierPriority {
	switch m {
	case "=":
		return ModifierExactly
	case "^~":
		return ModifierForward
	case "~", "~*":
		return ModifierRegular
	case "":
		return ModifierNone
	default:
		panic(errors.Wrapf(ErrSyntax, "unsupported modifier %s", m))
	}
}
