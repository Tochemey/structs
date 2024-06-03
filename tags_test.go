/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 Fatih Arslan
 * Copyright (c) 2024 Arsene Tochemey
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package structs

import "testing"

func TestParseTag_Name(t *testing.T) {
	tags := []struct {
		tag string
		has bool
	}{
		{"", false},
		{"name", true},
		{"name,opt", true},
		{"name , opt, opt2", false}, // has a single whitespace
		{", opt, opt2", false},
	}

	for _, tag := range tags {
		name, _ := parseTag(tag.tag)

		if (name != "name") && tag.has {
			t.Errorf("Parse tag should return name: %#v", tag)
		}
	}
}

func TestParseTag_Opts(t *testing.T) {
	tags := []struct {
		opts string
		has  bool
	}{
		{"name", false},
		{"name,opt", true},
		{"name , opt, opt2", false}, // has a single whitespace
		{",opt, opt2", true},
		{", opt3, opt4", false},
	}

	// search for "opt"
	for _, tag := range tags {
		_, opts := parseTag(tag.opts)

		if opts.Has("opt") != tag.has {
			t.Errorf("Tag opts should have opt: %#v", tag)
		}
	}
}
