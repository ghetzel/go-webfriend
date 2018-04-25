package webfriend

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func eval(script string) (map[string]interface{}, error) {
	scope, err := NewEnvironment(nil).EvaluateString(script)

	if err == nil {
		return scope.Data(), nil
	} else {
		return nil, err
	}
}

func TestAssignments(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`null`: nil,
		`a`:    1,
		`b`:    true,
		`c`:    "Test",
		`c1`:   "Test Test",
		`c2`:   "Test {c}",
		`d`:    3.14159,
		`e`: []interface{}{
			float64(1),
			true,
			"Test",
			3.14159,
			[]interface{}{
				float64(1),
				true,
				"Test",
				3.14159,
			},
			map[string]interface{}{
				`ok`: true,
			},
		},
		`f`: map[string]interface{}{
			`ok`: true,
		},
		`g`: `g`,
		`h`: `h`,
		`i`: `i`,
		`j`: `j`,
		`k`: nil,
		`l`: `l`,
		`m`: `m`,
		`o`: `o`,
		`p`: nil,
		`q`: `q`,
		`s`: `s`,
		`t`: true,
		`u`: nil,
		`v`: 1,
		`w`: true,
		`x`: "Test",
		`y`: 3.14159,
		`z`: []interface{}{
			float64(1),
			true,
			"Test",
			3.14159,
		},
		`z0`: 1,
		`z1`: true,
		`z2`: `Test`,
		`z3`: 3.14159,
		`aa`: 1,
		`bb`: true,
		`cc`: "Test",
		`dd`: 3.14159,
		`ee`: map[string]interface{}{
			`ok`: true,
			`always`: map[string]interface{}{
				`finishing`: map[string]interface{}{
					`each_others`: `sentences`,
				},
			},
		},
		`ee1`: map[string]interface{}{
			`finishing`: map[string]interface{}{
				`each_others`: `sentences`,
			},
		},
		`ee2`: map[string]interface{}{
			`each_others`: `sentences`,
		},
		`ee3`:  `sentences`,
		`ee4`:  nil,
		`ee5`:  true,
		`ee6`:  `sentences`,
		`ekey`: `always`,
		`ee7`:  `sentences`,
		`ee8`: map[string]interface{}{
			`ok`: true,
			`always`: map[string]interface{}{
				`finishing`: map[string]interface{}{
					`each_others`: `sandwiches`,
					`other`: map[string]interface{}{
						`stuff`: map[string]interface{}{
							`too`: true,
						},
					},
				},
			},
		},
		`put_1`: `test 1`,
		`put_2`: `test {a}`,
	}

	script := `# set variables of with values of every type
    $null = null
    $a = 1
    $b = true
    $c = "Test"
    $c1 = "Test {c}"
    $c2 = 'Test {c}'
    $d = 3.14159
    $e = [1, true, "Test", 3.14159, [1, true, "Test", 3.14159], {ok:true}]
    $f = {
        ok: true,
    }
    $g, $h, $i = "g", "h", "i"
    $j, $k = "j"
    $l, $m = ["l", "m", "n"]
    $o, $p = ["o"]
    $q, _, $s = ["q", "r", "s"]
    $t = $f.ok
    $u = $f.nonexistent
    $v, $w, $x, $y, $z = [1, true, "Test", 3.14159, [1, true, "Test", 3.14159]]
    $z0 = $z[0]
    $z1 = $z[1]
    $z2 = $z[2]
    $z3 = $z[3]
    # capture command results as variables, and also put a bunch of them on the same line
    put 1 -> $aa; put true -> $bb; put "Test" -> $cc; put 3.14159 -> $dd
    put {
        ok: true,
        always: {
            finishing: {
                each_others: "sentences",
            },
        },
    } -> $ee
    $ee1, $ee2 = $ee.always, $ee.always.finishing
    $ee3, $ee4 = [$ee.always.finishing.each_others, $ee.always.finishing.each_others.sandwiches]
    $ee5 = $ee['ok']
    $ee6 = $ee['always'].finishing['each_others']
    $ekey = 'always'
    $ee7 = $ee[$ekey].finishing['each_others']
    $ee8 = $ee
    $ee8.always['finishing'].each_others = 'sandwiches'
    $ee8.always['finishing'].other['stuff'].too = true
    put "test {a}" -> $put_1
    put 'test {a}' -> $put_2`

	actual, err := eval(script)

	assert.NoError(err)

	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestIfScopes(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`a`:             `top_a`,
		`b`:             `top_b`,
		`a_if`:          `top_a`,
		`b_if`:          `if_b`,
		`a_if_if`:       `if_if_a`,
		`b_if_if`:       `if_b`,
		`a_after_if_if`: `top_a`,
		`b_after_if_if`: `if_b`,
		`a_after_if`:    `top_a`,
		`b_after_if`:    `top_b`,
		`enter_if_val`:  51,
		`enter_el_val`:  61,
		`result`:        nil,
	}

	script := `
        $a             = "top_a"
        $b             = "top_b"
        $a_if          = null
        $b_if          = null
        $a_if_if       = null
        $b_if_if       = null
        $a_after_if_if = null
        $b_after_if_if = null
        $a_after_if    = null
        $b_after_if    = null

        if $b = "if_b"; $b {
            $a_if = $a
            $b_if = $b
            if $a = "if_if_a"; $a {
                $a_if_if = $a
                $b_if_if = $b
            }
            $a_after_if_if = $a
            $b_after_if_if = $b
        }
        $a_after_if = $a
        $b_after_if = $b
        $enter_if_val = null
        $enter_el_val = null

        # if condition trigger, verify condition value, populate via assignment
        if $value = 51; $value > 50 {
            $enter_if_val = 51
        } else {
            $enter_if_val = 9999
        }

        # else condition trigger, verify condition value, populate via command output
        if put 61 -> $value; $value > 100 {
            $enter_el_val = 7777
        } else {
            $enter_el_val = 61
        }
        $result = null`

	actual, err := eval(script)

	assert.NoError(err)

	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestConditionals(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`ten`:            10,
		`unset`:          nil,
		`true`:           true,
		`false`:          false,
		`string`:         "string",
		`names`:          []interface{}{"Bob", "Steve", "Fred"},
		`if_eq`:          true,
		`if_ne`:          true,
		`if_eq_null`:     true,
		`if_true`:        true,
		`if_false`:       true,
		`if_gt`:          true,
		`if_gte`:         true,
		`if_lt`:          true,
		`if_lte`:         true,
		`if_in`:          true,
		`if_not_in`:      true,
		`if_match_1`:     true,
		`if_match_2`:     true,
		`if_match_3`:     true,
		`if_not_match_1`: true,
		`if_not_match_2`: true,
		`if_not_match_3`: true,
		`if_match_4`:     true,
		`if_match_5`:     true,
		`if_match_6`:     true,
		`if_not_match_4`: true,
		`if_not_match_5`: true,
		`if_not_match_6`: true,
	}

	script := `
        $ten = 10
        $unset = null
        $true = true
        $false = false
        $string = "string"
        $names = ["Bob", "Steve", "Fred"]
        $if_eq = null
        $if_ne = null
        $if_eq_null = null
        $if_true = null
        $if_gt = null
        $if_gte = null
        $if_lt = null
        $if_lte = null
        $if_in = null
        $if_not_in = null
        $if_match_1 = null
        $if_match_2 = null
        $if_match_3 = null
        $if_match_4 = null
        $if_match_5 = null
        $if_match_6 = null
        $if_not_match_1 = null
        $if_not_match_2 = null
        $if_not_match_3 = null
        $if_not_match_4 = null
        $if_not_match_5 = null
        $if_not_match_6 = null
        if $ten == 10                    { $if_eq          = true }
        if $unset == null                { $if_eq_null     = true }
        if $ten != 5                     { $if_ne          = true }
        if $ten > 5                      { $if_gt          = true }
        if $ten >= 10                    { $if_gte         = true }
        if $ten < 20                     { $if_lt          = true }
        if $ten <= 10                    { $if_lte         = true }
        if $true                         { $if_true        = true }
        if not $false                    { $if_false       = true }
        if "Steve" in $names             { $if_in          = true }
        if "Bill" not in $names          { $if_not_in      = true }
        if $string =~ /str[aeiou]ng/     { $if_match_1     = true }
        if $string =~ /String/i          { $if_match_2     = true }
        if $string =~ /.*/               { $if_match_3     = true }
        if $string !~ /strong/i          { $if_not_match_1 = true }
        if $string !~ /String/           { $if_not_match_2 = true }
        if $string !~ /^ring$/           { $if_not_match_3 = true }
        if not $string !~ /str[aeiou]ng/ { $if_match_4     = true }
        if not $string !~ /String/i      { $if_match_5     = true }
        if not $string !~ /.*/           { $if_match_6     = true }
        if not $string =~ /strong/i      { $if_not_match_4 = true }
        if not $string =~ /String/       { $if_not_match_5 = true }
        if not $string =~ /^ring$/       { $if_not_match_6 = true }`

	actual, err := eval(script)
	assert.NoError(err)
	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestExpressions(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`a`:     2,
		`b`:     6,
		`c`:     20,
		`d`:     5,
		`aa`:    2,
		`bb`:    6,
		`cc`:    20,
		`dd`:    5,
		`f`:     `This 2 is {b} and done`,
		`put_a`: `    this is some stuff`,
		`put_b`: "    buncha\n    muncha\n    cruncha\n    lines",
	}

	script := `
        $a = 1 + 1
        $b = 9 - 3
        $c = 5 * 4
        $d = 50 / 10
        #$e = 4 * -6 * (3 * 7 + 5) + 2 * 7
        $aa = 1
        $aa += 1
        $bb = 9
        $bb -= 3
        $cc = 5
        $cc *= 4
        $dd = 50
        $dd /= 10
        $f = "This {a}" + ' is {b}' + " and done"

        put begin
            this is some stuff
        end -> $put_a
        put begin
            buncha
            muncha
            cruncha
            lines
        end -> $put_b`

	actual, err := eval(script)
	assert.NoError(err)
	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func TestLoops(t *testing.T) {
	assert := require.New(t)

	expected := map[string]interface{}{
		`forevers`:        9,
		`double_break`:    []interface{}{4, 1},
		`double_continue`: []interface{}{8, 9},
		`iterations`:      4,
		`things`: []interface{}{
			float64(1),
			float64(2),
			float64(3),
			float64(4),
			float64(5),
		},
		`topindex`: 9,
		`map`: map[string]interface{}{
			`first`:  float64(1),
			`second`: float64(2),
			`third`:  float64(3),
		},
		`m1`: `first:1`,
		`m2`: `second:2`,
		`m3`: `third:3`,
	}

	script := `
        $forevers = 0
        $double_break = null
        $double_continue = null
        $map = {
            first:  1,
            second: 2,
            third:  3,
        }
        $iterations = null
        $things = [1,2,3,4,5]
        loop {
            if not $index < 10 {
                break
            }
            $forevers = $index
        }
        loop $x in $things {
            $iterations = $index
        }
        loop count 10 {
            $topindex = $index
            loop count 10 {
                if $topindex == 4 {
                    if $index == 2 {
                        break 2
                    }
                }
                $double_break = [$topindex, $index]
            }
        }
        loop count 10 {
            $topindex = $index
            loop count 10 {
                if $topindex == 9 {
                    if $index >= 0 {
                        continue 2
                    }
                }
                $double_continue = [$topindex, $index]
            }
        }

        loop $k, $v in $map {
            if $index == 0 {
                $m1 = "{k}:{v}"
            } else if $index == 1 {
                $m2 = "{k}:{v}"
            } else {
                $m3 = "{k}:{v}"
            }
        }`

	actual, err := eval(script)
	assert.NoError(err)
	// fmt.Println(jsondiff(expected, actual))
	assert.Equal(expected, actual)
}

func jsondiff(expected interface{}, actual interface{}) string {
	if expectedJ, err := json.MarshalIndent(expected, ``, `  `); err == nil {
		if actualJ, err := json.MarshalIndent(actual, ``, `  `); err == nil {
			differ := gojsondiff.New()
			if d, err := differ.Compare(expectedJ, actualJ); err == nil {
				formatter := formatter.NewAsciiFormatter(expected, formatter.AsciiFormatterConfig{
					ShowArrayIndex: true,
					Coloring:       true,
				})

				if diffString, err := formatter.Format(d); err == nil {
					return diffString
				} else {
					return fmt.Sprintf("ERROR: formatter: %v", err)
				}
			} else {
				return fmt.Sprintf("ERROR: diff: %v", err)
			}
		} else {
			return fmt.Sprintf("ERROR: actual: %v", err)
		}
	} else {
		return fmt.Sprintf("ERROR: expected: %v", err)
	}
}
