---
title: Profile Rules
difficulty: easy
after: json-apis
hints:
  - 'Start with `errs := map[string]string{}` and check each field on its own, adding `errs["handle"] = "..."` when a rule fails. Don''t `return` early: the client wants to see **every** problem at once.'
  - '`len(s)` counts bytes, so `"🧀"` has length 4. Count characters with `utf8.RuneCountInString`. For the handle, loop over its runes and check each one is `a`-`z`, `0`-`9` or `_`.'
  - '`url.Parse` accepts almost anything, including `"pip.example"` (no scheme, no host). After parsing, check `u.Scheme == "https"` and `u.Host != ""` yourself.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    type ProfileUpdate struct {
    	Handle  string `json:"handle"`
    	Bio     string `json:"bio"`
    	Website string `json:"website"`
    }

    // validateProfile checks p and returns one message per invalid field,
    // keyed by the field's JSON name ("handle", "bio", "website").
    // It returns an empty (or nil) map if p is valid.
    //
    //   - handle: 3 to 15 characters, only lowercase letters, digits and _
    //   - bio: optional, at most 160 characters (characters, not bytes)
    //   - website: optional; if set, an https URL with a host
    func validateProfile(p ProfileUpdate) map[string]string {
    	// Check every field and collect all the problems, not just the first.
    	return nil
    }

    func main() {
    	fmt.Println(validateProfile(ProfileUpdate{Handle: "pip", Bio: "I like 🧀"}))
    	fmt.Println(validateProfile(ProfileUpdate{Handle: "Pip!", Website: "http://pip.example"}))
    	// want: map[]
    	// want: map[handle:... website:...]
    }
  solution: |
    package main

    import (
    	"fmt"
    	"net/url"
    	"unicode/utf8"
    )

    type ProfileUpdate struct {
    	Handle  string `json:"handle"`
    	Bio     string `json:"bio"`
    	Website string `json:"website"`
    }

    func validHandle(h string) bool {
    	if n := utf8.RuneCountInString(h); n < 3 || n > 15 {
    		return false
    	}
    	for _, r := range h {
    		if !('a' <= r && r <= 'z' || '0' <= r && r <= '9' || r == '_') {
    			return false
    		}
    	}
    	return true
    }

    func validateProfile(p ProfileUpdate) map[string]string {
    	errs := map[string]string{}
    	if !validHandle(p.Handle) {
    		errs["handle"] = "must be 3 to 15 characters: lowercase letters, digits or _"
    	}
    	if utf8.RuneCountInString(p.Bio) > 160 {
    		errs["bio"] = "must be at most 160 characters"
    	}
    	if p.Website != "" {
    		u, err := url.Parse(p.Website)
    		if err != nil || u.Scheme != "https" || u.Host == "" {
    			errs["website"] = "must be an https URL"
    		}
    	}
    	return errs
    }

    func main() {
    	fmt.Println(validateProfile(ProfileUpdate{Handle: "pip", Bio: "I like 🧀"}))
    	fmt.Println(validateProfile(ProfileUpdate{Handle: "Pip!", Website: "http://pip.example"}))
    }
  tests: |
    package main

    import (
    	"maps"
    	"slices"
    	"strings"
    	"testing"
    )

    func TestValidateProfile(t *testing.T) {
    	tests := []struct {
    		name string
    		p    ProfileUpdate
    		want []string // fields that must be reported, sorted
    	}{
    		{"minimal", ProfileUpdate{Handle: "pip"}, nil},
    		{"everything", ProfileUpdate{Handle: "big_cheese_2026", Bio: "I like 🧀", Website: "https://pip.example/about"}, nil},
    		{"handle too short", ProfileUpdate{Handle: "pi"}, []string{"handle"}},
    		{"handle too long", ProfileUpdate{Handle: "abcdefghijklmnop"}, []string{"handle"}},
    		{"empty handle", ProfileUpdate{}, []string{"handle"}},
    		{"uppercase handle", ProfileUpdate{Handle: "Pip"}, []string{"handle"}},
    		{"handle with space", ProfileUpdate{Handle: "pip squeak"}, []string{"handle"}},
    		{"handle with dash", ProfileUpdate{Handle: "pip-squeak"}, []string{"handle"}},
    		{"handle with accent", ProfileUpdate{Handle: "piné"}, []string{"handle"}},
    		{"bio of 160 emoji", ProfileUpdate{Handle: "pip", Bio: strings.Repeat("🧀", 160)}, nil},
    		{"bio of 161 letters", ProfileUpdate{Handle: "pip", Bio: strings.Repeat("a", 161)}, []string{"bio"}},
    		{"http website", ProfileUpdate{Handle: "pip", Website: "http://pip.example"}, []string{"website"}},
    		{"website without scheme", ProfileUpdate{Handle: "pip", Website: "pip.example"}, []string{"website"}},
    		{"website without host", ProfileUpdate{Handle: "pip", Website: "https:///about"}, []string{"website"}},
    		{"javascript website", ProfileUpdate{Handle: "pip", Website: "javascript:alert(1)"}, []string{"website"}},
    		{"garbled website", ProfileUpdate{Handle: "pip", Website: "https://pip example/%zz"}, []string{"website"}},
    		{"all wrong", ProfileUpdate{Handle: "P", Bio: strings.Repeat("b", 200), Website: "ftp://pip.example"}, []string{"bio", "handle", "website"}},
    	}
    	for _, tt := range tests {
    		errs := validateProfile(tt.p)
    		got := slices.Sorted(maps.Keys(errs))
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("%s: validateProfile(%+v) reported fields %q, want %q", tt.name, tt.p, got, tt.want)
    		}
    		for field, msg := range errs {
    			if msg == "" {
    				t.Errorf("%s: field %q has an empty message", tt.name, field)
    			}
    		}
    	}
    }
---

Mice can edit their Squeak profile with `PATCH /api/me`. The JSON decodes into a
`ProfileUpdate` just fine, but that doesn't make it *valid*. Rather than
rejecting the first bad field, Squeak answers with **all** of them, so the app
can highlight every problem at once:

```json
{"errors": {"handle": "must be 3 to 15 characters...", "website": "must be an https URL"}}
```

Complete `validateProfile(p)`. It returns a map from JSON field name to a
message for each rule that fails, and an empty or `nil` map when `p` is valid:

- **`handle`** (required): 3 to 15 characters, each a lowercase ASCII letter, a
  digit or `_`.
- **`bio`** (optional): at most 160 characters. Emoji count as one character
  each.
- **`website`** (optional): if it isn't empty, it must be an `https` URL with a
  host.

## Examples

```go
validateProfile(ProfileUpdate{Handle: "pip", Bio: "I like 🧀"})           // map[]
validateProfile(ProfileUpdate{Handle: "Pip!", Website: "pip.example"})  // map[handle:... website:...]
```

## Constraints

- The tests only check **which** fields are reported, and that every message is
  non-empty. The wording is yours.
- Report every failing field, not just the first.
