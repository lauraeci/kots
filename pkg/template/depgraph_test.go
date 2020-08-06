package template

import (
	"fmt"
	"testing"

	kotsv1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/kots/kotskinds/multitype"
	"github.com/stretchr/testify/require"
	"go.undefinedlabs.com/scopeagent"
)

type depGraphTestCase struct {
	dependencies    map[string][]string
	testCerts       map[string][]string
	testKeys        map[string][]string
	testCAs         map[string][]string
	testCAFromCerts map[string][][2]string
	testCAFromKeys  map[string][][2]string
	resolveOrder    []string
	expectError     bool     //expect an error fetching head nodes
	expectNotFound  []string //expect these dependencies not to be part of the head nodes

	name string
}

func TestDepGraph(t *testing.T) {
	tests := []depGraphTestCase{
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"bravo"},
				"delta":   {"alpha", "charlie"},
				"echo":    {},
			},
			resolveOrder: []string{"alpha", "bravo", "charlie", "delta", "echo"},
			name:         "basic_dependency_chain",
		},
		{
			dependencies: map[string][]string{
				"alpha": {"bravo"},
				"bravo": {"alpha"},
			},
			resolveOrder: []string{"alpha", "bravo"},
			expectError:  true,
			name:         "basic_circle",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"alpha"},
				"delta":   {"bravo", "charlie"},
				"echo":    {"delta"},
			},
			resolveOrder: []string{"alpha", "bravo", "charlie", "delta", "echo"},
			name:         "basic_forked_chain",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"alpha"},
				"delta":   {"bravo", "charlie", "foxtrot"},
				"echo":    {"delta"},
				"foxtrot": {},
			},
			resolveOrder:   []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot"},
			expectNotFound: []string{"delta"},
			name:           "unresolved_dependency",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {},
				"charlie": {"alpha"},
				"delta":   {"bravo"},
				"echo":    {"delta"},
			},
			resolveOrder: []string{"alpha", "bravo", "charlie", "delta", "echo"},
			name:         "two_chains",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"alpha", "bravo"},
				"delta":   {"alpha", "bravo", "charlie"},
				"echo":    {"alpha", "bravo", "charlie", "delta"},
				"foxtrot": {"alpha", "bravo", "charlie", "delta", "echo"},
				"golf":    {"alpha", "bravo", "charlie", "delta", "echo", "foxtrot"},
				"hotel":   {"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf"},
				"india":   {"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel"},
				"juliet":  {"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india"},
				"kilo":    {"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet"},
				"lima":    {"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet", "kilo"},
				"mike":    {"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet", "kilo", "lima"},
			},
			resolveOrder: []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet", "kilo", "lima", "mike"},
			name:         "pyramid",
		},
		{
			dependencies: map[string][]string{
				"alpha": {},
				"bravo": {},
			},
			resolveOrder: []string{"alpha", "bravo"},
			expectError:  false,
			name:         "not referenced", // items should eventually be resolved even if nothing depends on them
		},
		{
			dependencies: map[string][]string{
				"alpha": {},
				"bravo": {"alpha", "charlie"},
			},
			resolveOrder: []string{"alpha", "bravo"},
			expectError:  true,
			name:         "does_not_exist",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"bravo"},
				"delta":   {"alpha", "charlie"},
				"echo":    {},
			},
			testCerts: map[string][]string{
				"echo": {"certA"},
			},
			testKeys: map[string][]string{
				"delta": {"certA"},
			},
			resolveOrder: []string{"alpha", "bravo", "charlie", "echo", "delta"},
			name:         "basic_certs",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"bravo"},
				"delta":   {"alpha", "charlie"},
				"echo":    {},
			},
			testCerts: map[string][]string{
				"echo": {"certA"},
			},
			testKeys: map[string][]string{
				"delta": {"certA"},
			},
			resolveOrder:   []string{"alpha", "bravo", "charlie", "delta", "echo"},
			name:           "basic_certs_original_order",
			expectNotFound: []string{"delta"},
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"alpha", "bravo"},
				"delta":   {},
				"echo":    {},
			},
			testCAs: map[string][]string{
				"echo": {"caA"},
			},
			testCAFromCerts: map[string][][2]string{
				"delta": {{"caA", "certA"}},
			},
			testCAFromKeys: map[string][][2]string{
				"charlie": {{"caA", "certA"}},
			},
			resolveOrder: []string{"alpha", "bravo", "echo", "delta", "charlie"},
			name:         "basic_cacerts",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"alpha", "bravo"},
				"delta":   {},
				"echo":    {},
			},
			testCAs: map[string][]string{
				"echo": {"caA"},
			},
			testCAFromCerts: map[string][][2]string{
				"delta": {{"caA", "certA"}},
			},
			testCAFromKeys: map[string][][2]string{
				"charlie": {{"caA", "certA"}},
			},
			resolveOrder:   []string{"alpha", "bravo", "charlie", "delta", "echo"},
			name:           "basic_cacerts_original_order",
			expectNotFound: []string{"delta", "charlie"},
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"bravo"},
				"delta":   {"alpha", "charlie"},
				"echo":    {},
			},
			testCAs: map[string][]string{
				"echo": {"caA"},
			},
			testCAFromKeys: map[string][][2]string{
				"delta": {{"caA", "certA"}},
			},
			resolveOrder: []string{"alpha", "bravo", "charlie", "echo", "delta"},
			name:         "basic_cacerts_key_only",
		},
		{
			dependencies: map[string][]string{
				"alpha":   {},
				"bravo":   {"alpha"},
				"charlie": {"bravo"},
				"delta":   {"alpha", "charlie"},
				"echo":    {},
			},
			testCAs: map[string][]string{
				"echo": {"caA"},
			},
			testCAFromKeys: map[string][][2]string{
				"delta": {{"caA", "certA"}},
			},
			resolveOrder:   []string{"alpha", "bravo", "charlie", "delta", "echo"},
			name:           "basic_cacerts_key_only_original_order",
			expectNotFound: []string{"delta"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scopetest := scopeagent.StartTest(t)
			defer scopetest.End()

			graph := depGraph{}
			for source, deps := range test.dependencies {
				graph.AddNode(source)
				for _, dep := range deps {
					graph.AddDep(source, dep)
				}
			}
			for source, certNames := range test.testCerts {
				for _, certName := range certNames {
					graph.AddCert(source, certName)
				}
			}
			for source, keys := range test.testKeys {
				for _, key := range keys {
					graph.AddKey(source, key)
				}
			}
			for source, caNames := range test.testCAs {
				for _, caName := range caNames {
					graph.AddCA(source, caName)
				}
			}
			for source, caCertNames := range test.testCAFromCerts {
				for _, caCertName := range caCertNames {
					graph.AddCertFromCA(source, caCertName[0], caCertName[1])
				}
			}
			for source, caCertNames := range test.testCAFromKeys {
				for _, caCertName := range caCertNames {
					graph.AddKeyFromCA(source, caCertName[0], caCertName[1])
				}
			}
			graph.resolveCertKeys()
			graph.resolveCACerts()
			graph.resolveCACertKeys()

			runGraphTests(t, test, graph)
		})

		t.Run(test.name+"+parse", func(t *testing.T) {
			scopetest := scopeagent.StartTest(t)
			defer scopetest.End()

			graph := depGraph{}

			groups := buildTestConfigGroups(
				test.dependencies, test.testCerts, test.testKeys,
				test.testCAs, test.testCAFromCerts, test.testCAFromKeys,
				"templateStringStart", "templateStringEnd", true,
			)

			err := graph.ParseConfigGroup(groups)
			require.NoError(t, err)

			runGraphTests(t, test, graph)
		})
	}
}

// this makes sure that we test with each of the configOption types, in both Value and Default
func buildTestConfigGroups(dependencies, certs, keys, cas map[string][]string, caFromCerts, caFromKeys map[string][][2]string, prefix string, suffix string, rotate bool) []kotsv1beta1.ConfigGroup {
	group := kotsv1beta1.ConfigGroup{}
	group.Items = make([]kotsv1beta1.ConfigItem, 0)
	counter := 0

	templateFuncs := []string{
		"{{repl ConfigOption \"%s\" }}",
		"{{repl ConfigOptionIndex \"%s\" }}",
		"{{repl ConfigOptionData \"%s\" }}",
		"repl{{ ConfigOptionEquals \"%s\" \"abc\" }}",
		"{{repl ConfigOptionNotEquals \"%s\" \"xyz\" }}",
	}

	if !rotate {
		//use only ConfigOption, not all 5
		templateFuncs = []string{
			"{{repl ConfigOption \"%s\" }}",
		}
	}

	totalDepItems := 0

	for source, deps := range dependencies {
		newItem := kotsv1beta1.ConfigItem{Type: "text", Name: source}
		depString := prefix
		for _, dep := range deps {
			depString += fmt.Sprintf(templateFuncs[totalDepItems%len(templateFuncs)], dep)
			totalDepItems++
		}

		if certNames, ok := certs[source]; ok {
			for _, certName := range certNames {
				depString += fmt.Sprintf("{{repl TLSCert \"%s\" }}", certName)
			}
		}

		if certNames, ok := keys[source]; ok {
			for _, certName := range certNames {
				depString += fmt.Sprintf("{{repl TLSKey \"%s\" }}", certName)
			}
		}

		if caNames, ok := cas[source]; ok {
			for _, caName := range caNames {
				depString += fmt.Sprintf("{{repl TLSCACert \"%s\" }}", caName)
			}
		}

		if caCertNames, ok := caFromCerts[source]; ok {
			for _, caCertName := range caCertNames {
				depString += fmt.Sprintf("{{repl TLSCertFromCA \"%s\" \"%s\" }}", caCertName[0], caCertName[1])
			}
		}

		if caCertNames, ok := caFromKeys[source]; ok {
			for _, caCertName := range caCertNames {
				depString += fmt.Sprintf("{{repl TLSKeyFromCA \"%s\" \"%s\" }}", caCertName[0], caCertName[1])
			}
		}

		depString += suffix

		if counter%2 == 0 {
			newItem.Value.StrVal = depString
			newItem.Value.Type = multitype.String
		} else {
			newItem.Default.StrVal = depString
			newItem.Default.Type = multitype.String
		}
		counter++

		group.Items = append(group.Items, newItem)
	}

	return []kotsv1beta1.ConfigGroup{group}
}

func runGraphTests(t *testing.T, test depGraphTestCase, graph depGraph) {
	depLen := len(graph.Dependencies)
	graphCopy, err := graph.Copy()
	require.NoError(t, err)

	for _, toResolve := range test.resolveOrder {
		available, err := graph.GetHeadNodes()
		if err != nil && test.expectError {
			// fmt.Printf("err: %s\n", err.Error())
			return
		}

		require.NoError(t, err, "toResolve: %s", toResolve)

		if stringInSlice(toResolve, test.expectNotFound) {
			require.NotContains(t, available, toResolve)
			return
		}

		require.Contains(t, available, toResolve)

		graph.ResolveDep(toResolve)
	}

	available, err := graph.GetHeadNodes()
	require.NoError(t, err)
	require.Empty(t, available)

	require.False(t, test.expectError, "Did not find expected error")

	require.Equal(t, depLen, len(graphCopy.Dependencies))
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
