package vulnmap

import (
	"strings"
	"testing"

	"ariadne/internal/mavenversion"
)

// TestReleaseDiffManualMappings pins the 22 manual entries of 2026-11-13
// (advisories with no usable fix commit and no class name, mapped from a
// hand-read release diff): every release whose bytecode was checked is
// covered by exactly one variant with the expected methods, and the
// fixing release (or a version below the checked span) selects nothing —
// MAPPING_VERSION_UNCOVERED, never a guess.
func TestReleaseDiffManualMappings(t *testing.T) {
	meta, err := LoadManualMeta("")
	if err != nil {
		t.Fatal(err)
	}
	pick := func(id, pkg, v string) []string {
		var out []string
		n := 0
		for _, vr := range meta[id].Variants {
			if vr.Package == pkg && vr.Range.Contains(v) == mavenversion.In {
				n++
				out = append(out, vr.Methods...)
			}
		}
		if n > 1 {
			t.Errorf("%s %s %s: %d variants cover it", id, pkg, v, n)
		}
		return out
	}
	const (
		yaml  = "org.yaml:snakeyaml"
		mvc   = "org.springframework:spring-webmvc"
		spel  = "org.springframework:spring-expression"
		core  = "org.springframework:spring-core"
		th    = "org.thymeleaf:thymeleaf"
		thsp5 = "org.thymeleaf:thymeleaf-spring5"
		R     = "org.springframework.web.servlet.resource."
		SSE   = "org.springframework.web.servlet.mvc.method.annotation.SseEmitter$SseEventBuilderImpl#"
		E     = "org.springframework.expression.spel."
		SU    = "org.thymeleaf.spring5.util.SpringStandardExpressionUtils#"
		CU    = "org.thymeleaf.standard.util.StandardExpressionUtils#"
	)
	composer := []string{"org.yaml.snakeyaml.composer.Composer#getSingleNode", "org.yaml.snakeyaml.composer.Composer#getNode"}
	yamlChecked := []string{"1.17", "1.23", "1.26", "1.27", "1.28", "1.29", "1.30"}
	mvcChecked := []string{"5.3.0", "5.3.4", "5.3.6", "5.3.20", "5.3.39", "6.0.0", "6.0.23", "6.1.0", "6.1.21", "6.2.0"}
	spelChecked := []string{"4.3.0.RELEASE", "4.3.4.RELEASE", "4.3.30.RELEASE", "5.0.0.RELEASE", "5.2.25.RELEASE", "5.3.0", "5.3.4", "5.3.6", "5.3.39", "6.0.0", "6.1.21", "6.2.0", "6.2.18"}
	logChecked := []string{"5.2.0.RELEASE", "5.2.17.RELEASE", "5.3.0", "5.3.4", "5.3.6", "5.3.10"}
	rel := func(vs ...string) []string {
		out := make([]string, len(vs))
		for i, v := range vs {
			out[i] = v + ".RELEASE"
		}
		return out
	}
	thOld, thNew := rel("3.0.12", "3.0.13"), rel("3.0.14", "3.0.15", "3.1.0", "3.1.3")
	type span struct {
		pkg      string
		role     string
		versions []string // bytecode checked
		outside  []string // must select nothing
		methods  []string
	}
	cases := map[string][]span{
		"CVE-2022-38749": {{yaml, RoleEntrypoint, yamlChecked, []string{"1.31", "1.16"}, composer}},
		"CVE-2022-38750": {{yaml, RoleEntrypoint, yamlChecked, []string{"1.31"}, composer}},
		"CVE-2022-38751": {{yaml, RoleEntrypoint, yamlChecked, []string{"1.31"}, composer}},
		"CVE-2022-38752": {{yaml, RoleEntrypoint, append(yamlChecked, "1.31"), []string{"1.32"}, composer}},
		"CVE-2022-41854": {{yaml, RoleEntrypoint, append(yamlChecked, "1.31"), []string{"1.32"}, composer}},
		"CVE-2026-22735": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.16"), []string{"6.2.17", "5.3.40", "4.3.30.RELEASE"},
			[]string{SSE + "id", SSE + "name", SSE + "comment", SSE + "data", SSE + "append", SSE + "saveAppendedText", SSE + "build"}}},
		"CVE-2026-22737": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.16"), []string{"6.2.17"},
			[]string{"org.springframework.web.servlet.view.script.ScriptTemplateView#getResource"}}},
		"CVE-2026-22741": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.17"), []string{"6.2.18", "5.3.47"},
			[]string{R + "CachingResourceResolver#getContentCodingKey", R + "EncodedResourceResolver#resolveResourceInternal", R + "EncodedResourceResolver#getAcceptEncoding"}}},
		"CVE-2026-22745": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.17"), []string{"6.2.18"},
			[]string{R + "ResourceHttpRequestHandler#getResource", R + "ResourceHttpRequestHandler#handleRequest"}}},
		"CVE-2026-41841": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.18"), []string{"6.2.19"},
			[]string{R + "CachingResourceResolver#computeKey", R + "CachingResourceResolver#resolveResourceInternal"}}},
		"CVE-2026-41842": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.18"), []string{"6.2.19"},
			[]string{R + "VersionResourceResolver#resolveResourceInternal", R + "AbstractVersionStrategy$FileNameVersionPathStrategy#removeVersion"}}},
		"CVE-2026-41843": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.18"), []string{"6.2.19"},
			[]string{R + "VersionResourceResolver#resolveResourceInternal", R + "AbstractVersionStrategy$FileNameVersionPathStrategy#removeVersion"}}},
		"CVE-2026-41853": {{mvc, RoleVulnerable, append(mvcChecked, "6.2.18"), []string{"6.2.19"},
			[]string{"org.springframework.http.codec.multipart.MultipartParser$HeadersState#onNext", "org.springframework.http.codec.multipart.DefaultPartHttpMessageReader#read"}}},
		"CVE-2026-41849": {{spel, RoleVulnerable, spelChecked, []string{"6.2.19", "4.2.9.RELEASE"},
			[]string{E + "ast.OpMultiply#getValueInternal", E + "ast.OpPlus#getValueInternal"}}},
		"CVE-2026-41850": {{spel, RoleEntrypoint, spelChecked, []string{"6.2.19"}, []string{E + "standard.SpelExpression#getValue"}}},
		"CVE-2026-41852": {{spel, RoleVulnerable, spelChecked, []string{"6.2.19"}, []string{E + "support.ReflectivePropertyAccessor#findGetterForProperty"}}},
		"CVE-2021-22096": {{core, RoleVulnerable, logChecked, []string{"5.3.11", "5.2.18.RELEASE", "5.1.20.RELEASE"},
			[]string{"org.springframework.core.log.LogFormatUtils#formatValue"}}},
		"CVE-2021-22060": {{core, RoleVulnerable, append(logChecked, "5.2.18.RELEASE", "5.3.11", "5.3.13"), []string{"5.3.14", "5.2.19.RELEASE"},
			[]string{"org.springframework.core.log.LogFormatUtils#formatValue"}}},
		"CVE-2021-43466": {{thsp5, RoleVulnerable, rel("3.0.12"), rel("3.0.13", "3.0.11"),
			[]string{"org.thymeleaf.spring5.util.SpringRequestUtils#checkViewNameNotInRequest", SU + "containsSpELInstantiationOrStatic"}}},
		"CVE-2026-40477": {
			{th, RoleVulnerable, append(rel("3.0.11"), append(thOld, thNew...)...), rel("3.1.4"), []string{
				"org.thymeleaf.standard.expression.OGNLVariableExpressionEvaluator#evaluate", "org.thymeleaf.standard.expression.OGNLExpressionObjectsWrapper#get",
				"org.thymeleaf.standard.expression.OGNLContextPropertyAccessor#getProperty", "org.thymeleaf.standard.expression.OGNLShortcutExpression#getContextProperty"}},
			{thsp5, RoleVulnerable, append(rel("3.0.11"), append(thOld, thNew...)...), rel("3.1.4"), []string{
				"org.thymeleaf.spring5.expression.SPELVariableExpressionEvaluator#evaluate", "org.thymeleaf.spring5.expression.ThymeleafEvaluationContext#lookupVariable",
				"org.thymeleaf.spring5.expression.ThymeleafEvaluationContextWrapper#lookupVariable", "org.thymeleaf.spring5.expression.SPELContextMapWrapper#containsKey"}},
		},
		"CVE-2026-40478": {
			{th, RoleVulnerable, thOld, rel("3.1.4", "3.0.11"), []string{CU + "containsOGNLInstantiationOrStatic", "org.thymeleaf.standard.expression.FragmentExpression#indexOfLastParenthesesGroup"}},
			{th, RoleVulnerable, thNew, nil, []string{CU + "containsOGNLInstantiationOrStaticOrParam", "org.thymeleaf.standard.expression.FragmentExpression#indexOfLastParenthesesGroup"}},
			{thsp5, RoleVulnerable, thOld, rel("3.1.4", "3.0.11"), []string{SU + "containsSpELInstantiationOrStatic", "org.thymeleaf.spring5.util.SpringRequestUtils#checkViewNameNotInRequest"}},
			{thsp5, RoleVulnerable, thNew, nil, []string{SU + "containsSpELInstantiationOrStaticOrParam", "org.thymeleaf.spring5.util.SpringRequestUtils#checkViewNameNotInRequest"}},
		},
		"CVE-2026-41901": {
			{th, RoleVulnerable, thOld, rel("3.1.5", "3.0.11"), []string{CU + "containsOGNLInstantiationOrStatic", "org.thymeleaf.standard.processor.AbstractStandardFragmentInsertionTagProcessor#doProcess"}},
			{th, RoleVulnerable, thNew, nil, []string{CU + "containsOGNLInstantiationOrStaticOrParam", "org.thymeleaf.standard.processor.AbstractStandardFragmentInsertionTagProcessor#doProcess"}},
			{th, RoleVulnerable, rel("3.1.4"), nil, []string{CU + "containsExternalAccess", "org.thymeleaf.standard.processor.AbstractStandardFragmentInsertionTagProcessor#doProcess"}},
			{thsp5, RoleVulnerable, thOld, rel("3.1.5"), []string{SU + "containsSpELInstantiationOrStatic"}},
			{thsp5, RoleVulnerable, thNew, nil, []string{SU + "containsSpELInstantiationOrStaticOrParam"}},
			{thsp5, RoleVulnerable, rel("3.1.4"), nil, []string{SU + "containsExternalAccess"}},
		},
	}
	if len(cases) != 22 {
		t.Fatalf("%d cases, want 22", len(cases))
	}
	for id, spans := range cases {
		m := meta[id]
		for _, s := range spans {
			if m.Role != s.role || m.Coverage != CoveragePartial || m.Review != ReviewReviewed {
				t.Errorf("%s: %s/%s/%s, want %s/partial/reviewed", id, m.Role, m.Coverage, m.Review, s.role)
			}
			for _, v := range s.versions {
				if got := pick(id, s.pkg, v); strings.Join(got, ",") != strings.Join(s.methods, ",") {
					t.Errorf("%s %s %s = %v, want %v", id, s.pkg, v, got, s.methods)
				}
			}
			for _, v := range s.outside {
				if got := pick(id, s.pkg, v); len(got) != 0 {
					t.Errorf("%s %s %s outside the checked span = %v", id, s.pkg, v, got)
				}
			}
		}
	}
}
