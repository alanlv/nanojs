package stdlib

import (
	"regexp"

	"github.com/zeaphoo/nanojs/v2"
)

func makeTextRegexp(re *regexp.Regexp) *nanojs.ImmutableMap {
	return &nanojs.ImmutableMap{
		Value: map[string]nanojs.Object{
			// match(text) => bool
			"match": &nanojs.UserFunction{
				Value: func(args ...nanojs.Object) (
					ret nanojs.Object,
					err error,
				) {
					if len(args) != 1 {
						err = nanojs.ErrWrongNumArguments
						return
					}

					s1, ok := nanojs.ToString(args[0])
					if !ok {
						err = nanojs.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					if re.MatchString(s1) {
						ret = nanojs.TrueValue
					} else {
						ret = nanojs.FalseValue
					}

					return
				},
			},

			// find(text) 			=> array(array({text:,begin:,end:}))/undefined
			// find(text, maxCount) => array(array({text:,begin:,end:}))/undefined
			"find": &nanojs.UserFunction{
				Value: func(args ...nanojs.Object) (
					ret nanojs.Object,
					err error,
				) {
					numArgs := len(args)
					if numArgs != 1 && numArgs != 2 {
						err = nanojs.ErrWrongNumArguments
						return
					}

					s1, ok := nanojs.ToString(args[0])
					if !ok {
						err = nanojs.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					if numArgs == 1 {
						m := re.FindStringSubmatchIndex(s1)
						if m == nil {
							ret = nanojs.UndefinedValue
							return
						}

						arr := &nanojs.Array{}
						for i := 0; i < len(m); i += 2 {
							arr.Value = append(arr.Value,
								&nanojs.ImmutableMap{
									Value: map[string]nanojs.Object{
										"text": &nanojs.String{
											Value: s1[m[i]:m[i+1]],
										},
										"begin": &nanojs.Int{
											Value: int64(m[i]),
										},
										"end": &nanojs.Int{
											Value: int64(m[i+1]),
										},
									}})
						}

						ret = &nanojs.Array{Value: []nanojs.Object{arr}}

						return
					}

					i2, ok := nanojs.ToInt(args[1])
					if !ok {
						err = nanojs.ErrInvalidArgumentType{
							Name:     "second",
							Expected: "int(compatible)",
							Found:    args[1].TypeName(),
						}
						return
					}
					m := re.FindAllStringSubmatchIndex(s1, i2)
					if m == nil {
						ret = nanojs.UndefinedValue
						return
					}

					arr := &nanojs.Array{}
					for _, m := range m {
						subMatch := &nanojs.Array{}
						for i := 0; i < len(m); i += 2 {
							subMatch.Value = append(subMatch.Value,
								&nanojs.ImmutableMap{
									Value: map[string]nanojs.Object{
										"text": &nanojs.String{
											Value: s1[m[i]:m[i+1]],
										},
										"begin": &nanojs.Int{
											Value: int64(m[i]),
										},
										"end": &nanojs.Int{
											Value: int64(m[i+1]),
										},
									}})
						}

						arr.Value = append(arr.Value, subMatch)
					}

					ret = arr

					return
				},
			},

			// replace(src, repl) => string
			"replace": &nanojs.UserFunction{
				Value: func(args ...nanojs.Object) (
					ret nanojs.Object,
					err error,
				) {
					if len(args) != 2 {
						err = nanojs.ErrWrongNumArguments
						return
					}

					s1, ok := nanojs.ToString(args[0])
					if !ok {
						err = nanojs.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					s2, ok := nanojs.ToString(args[1])
					if !ok {
						err = nanojs.ErrInvalidArgumentType{
							Name:     "second",
							Expected: "string(compatible)",
							Found:    args[1].TypeName(),
						}
						return
					}

					s, ok := doTextRegexpReplace(re, s1, s2)
					if !ok {
						return nil, nanojs.ErrStringLimit
					}

					ret = &nanojs.String{Value: s}

					return
				},
			},

			// split(text) 			 => array(string)
			// split(text, maxCount) => array(string)
			"split": &nanojs.UserFunction{
				Value: func(args ...nanojs.Object) (
					ret nanojs.Object,
					err error,
				) {
					numArgs := len(args)
					if numArgs != 1 && numArgs != 2 {
						err = nanojs.ErrWrongNumArguments
						return
					}

					s1, ok := nanojs.ToString(args[0])
					if !ok {
						err = nanojs.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "string(compatible)",
							Found:    args[0].TypeName(),
						}
						return
					}

					var i2 = -1
					if numArgs > 1 {
						i2, ok = nanojs.ToInt(args[1])
						if !ok {
							err = nanojs.ErrInvalidArgumentType{
								Name:     "second",
								Expected: "int(compatible)",
								Found:    args[1].TypeName(),
							}
							return
						}
					}

					arr := &nanojs.Array{}
					for _, s := range re.Split(s1, i2) {
						arr.Value = append(arr.Value,
							&nanojs.String{Value: s})
					}

					ret = arr

					return
				},
			},
		},
	}
}

// Size-limit checking implementation of regexp.ReplaceAllString.
func doTextRegexpReplace(re *regexp.Regexp, src, repl string) (string, bool) {
	idx := 0
	out := ""
	for _, m := range re.FindAllStringSubmatchIndex(src, -1) {
		var exp []byte
		exp = re.ExpandString(exp, repl, src, m)
		if len(out)+m[0]-idx+len(exp) > nanojs.MaxStringLen {
			return "", false
		}
		out += src[idx:m[0]] + string(exp)
		idx = m[1]
	}
	if idx < len(src) {
		if len(out)+len(src)-idx > nanojs.MaxStringLen {
			return "", false
		}
		out += src[idx:]
	}
	return out, true
}
