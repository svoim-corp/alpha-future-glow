// @ts-check
import js from '@eslint/js';
import stylistic from '@stylistic/eslint-plugin';
import jsxA11y from 'eslint-plugin-jsx-a11y';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import simpleImportSort from 'eslint-plugin-simple-import-sort';
import unusedImports from 'eslint-plugin-unused-imports';
import globals from 'globals';
import tseslint from 'typescript-eslint';

/* ------------------------------------------------------------------ */
/*  FSD: границы слоёв                                                  */
/*  Порядок сверху вниз. Слой может импортировать только слои НИЖЕ себя. */
/* ------------------------------------------------------------------ */
const LAYERS = ['app', 'pages', 'widgets', 'features', 'entities', 'shared'];
const SLICED_LAYERS = ['pages', 'widgets', 'features', 'entities'];

/**
 * Публичный API shared: @shared/ui, @shared/api, @shared/config, @shared/lib/<модуль>.
 * Всё остальное (голый @shared, глубокие пути) — запрещено.
 */
const SHARED_FORBIDDEN_REGEX = '^@shared(?!/(ui|api|config|lib/[^/]+)$)(/.*)?$';

const REDUX_HOOKS_RESTRICTION = {
	name: 'react-redux',
	importNames: ['useDispatch', 'useSelector', 'useStore'],
	message: 'Используй типизированные useAppDispatch / useAppSelector из @shared/lib/store.',
};

/**
 * @param {string} layer
 * @param {{ isTest: boolean }} options
 */
const fsdImportRules = (layer, { isTest }) => {
	const index = LAYERS.indexOf(layer);
	const upper = LAYERS.slice(0, index);
	const lower = LAYERS.slice(index + 1);

	/** @type {Array<Record<string, unknown>>} */
	const patterns = [];

	if (upper.length > 0) {
		patterns.push({
			regex: `^@(${upper.join('|')})(/|$)`,
			message: `FSD: слой «${layer}» не может импортировать вышележащие слои (${upper.join(', ')}).`,
		});
	}

	if (layer !== 'app') {
		patterns.push({
			regex: `^@${layer}(/|$)`,
			message: `FSD: внутри слоя «${layer}» импорт через алиас запрещён. `
				+ 'Внутри своего слайса — относительные пути; импорт другого слайса того же слоя запрещён.',
		});
	}

	for (const l of lower) {
		if (l === 'shared') {
			patterns.push({
				regex: SHARED_FORBIDDEN_REGEX,
				message: 'FSD: импортируй shared только через публичный API сегмента: '
					+ '@shared/ui, @shared/api, @shared/config, @shared/lib/<модуль>.',
			});
		} else if (SLICED_LAYERS.includes(l)) {
			patterns.push({
				// запрещены голый @layer и @layer/<slice>/<что-угодно>
				regex: `^@${l}(/[^/]+/.+)?$`,
				message: `FSD: импортируй слайс слоя «${l}» только через его public API: @${l}/<slice>.`,
			});
		}
	}

	patterns.push({
		regex: '^(\\.\\./)+(app|pages|widgets|features|entities|shared)(/|$)',
		message: 'FSD: не выходи из слайса относительным путём — используй алиас слоя и public API.',
	});

	if (!isTest) {
		patterns.push({
			regex: '^@test(/|$)',
			message: 'Тестовая инфраструктура (@test) доступна только в *.test.ts(x).',
		});
	}

	return {
		'no-restricted-imports': ['error', { paths: [REDUX_HOOKS_RESTRICTION], patterns }],
	};
};

const fsdConfigs = LAYERS.flatMap((layer) => [
	{
		name: `fsd/${layer}`,
		files: [`src/${layer}/**/*.{ts,tsx}`],
		ignores: ['**/*.test.{ts,tsx}'],
		rules: fsdImportRules(layer, { isTest: false }),
	},
	{
		name: `fsd/${layer}/tests`,
		files: [`src/${layer}/**/*.test.{ts,tsx}`],
		rules: fsdImportRules(layer, { isTest: true }),
	},
]);

export default tseslint.config(
	{
		ignores: ['dist', 'coverage', 'node_modules', 'public'],
	},

	js.configs.recommended,
	...tseslint.configs.recommendedTypeChecked,
	{
		languageOptions: {
			ecmaVersion: 'latest',
			sourceType: 'module',
			globals: { ...globals.browser },
			parserOptions: {
				projectService: true,
				tsconfigRootDir: import.meta.dirname,
			},
		},
	},
	{
		files: ['*.{js,ts}'],
		languageOptions: { globals: { ...globals.node } },
	},

	/* ---------- React ---------- */
	{
		files: ['src/**/*.{ts,tsx}', 'test/**/*.{ts,tsx}'],
		...react.configs.flat.recommended,
		...react.configs.flat['jsx-runtime'],
		settings: { react: { version: 'detect' } },
	},
	reactHooks.configs.flat['recommended-latest'],
	{
		files: ['src/**/*.tsx'],
		...jsxA11y.flatConfigs.recommended,
	},

	/* ---------- Общие правила ---------- */
	{
		plugins: {
			'@stylistic': stylistic,
			'simple-import-sort': simpleImportSort,
			'unused-imports': unusedImports,
		},
		rules: {
			// Импорты
			'simple-import-sort/imports': 'error',
			'simple-import-sort/exports': 'error',
			'unused-imports/no-unused-imports': 'error',
			'@typescript-eslint/no-unused-vars': 'off',
			'unused-imports/no-unused-vars': ['error', {
				vars: 'all',
				varsIgnorePattern: '^_',
				args: 'after-used',
				argsIgnorePattern: '^_',
			}],
			'@typescript-eslint/consistent-type-imports': ['error', { fixStyle: 'inline-type-imports' }],

			// TypeScript
			'@typescript-eslint/no-explicit-any': 'error',
			'@typescript-eslint/no-non-null-assertion': 'error',
			'@typescript-eslint/switch-exhaustiveness-check': 'error',
			'@typescript-eslint/naming-convention': ['error',
				{ selector: 'interface', format: ['PascalCase'] },
				{ selector: 'typeAlias', format: ['PascalCase'] },
				{ selector: 'class', format: ['PascalCase'] },
				{ selector: 'enum', format: ['PascalCase'] },
			],

			// Общие
			'no-console': ['error', { allow: ['warn', 'error'] }],
			'no-debugger': 'error',
			'eqeqeq': ['error', 'always'],
			'prefer-const': 'error',
			'no-var': 'error',

			// Стиль (как в wakadoo-web: табы, одинарные кавычки, точка с запятой)
			'@stylistic/indent': ['error', 'tab', { SwitchCase: 1 }],
			'@stylistic/jsx-indent-props': ['error', 'tab'],
			'@stylistic/quotes': ['error', 'single', { avoidEscape: true }],
			'@stylistic/jsx-quotes': ['error', 'prefer-double'],
			'@stylistic/semi': ['error', 'always'],
			'@stylistic/comma-dangle': ['error', 'always-multiline'],
			'@stylistic/object-curly-spacing': ['error', 'always'],
			'@stylistic/array-bracket-spacing': ['error', 'never'],
			'@stylistic/arrow-parens': ['error', 'always'],
			'@stylistic/operator-linebreak': ['error', 'before', { overrides: { '=': 'ignore' } }],
			'@stylistic/eol-last': ['error', 'always'],
			'@stylistic/no-trailing-spaces': 'error',
			'@stylistic/no-multiple-empty-lines': ['error', { max: 1, maxEOF: 0 }],
			'@stylistic/no-multi-spaces': 'error',
			'@stylistic/max-len': ['error', {
				code: 120,
				tabWidth: 4,
				ignoreUrls: true,
				ignoreStrings: true,
				ignoreTemplateLiterals: true,
				ignoreComments: true,
			}],
		},
	},

	/* ---------- Код приложения (src) ---------- */
	{
		files: ['src/**/*.{ts,tsx}'],
		rules: {
			'react/function-component-definition': ['error', {
				namedComponents: 'arrow-function',
				unnamedComponents: 'arrow-function',
			}],
			'react/jsx-no-useless-fragment': 'error',
			'react/self-closing-comp': 'error',
			'react/no-array-index-key': 'error',
			'react/jsx-boolean-value': ['error', 'never'],
			'react/prop-types': 'off',
			// Только именованные экспорты — проще контролировать public API слайсов
			'no-restricted-syntax': ['error', {
				selector: 'ExportDefaultDeclaration',
				message: 'Используй именованный экспорт.',
			}],
		},
	},

	...fsdConfigs,

	// Единственное место, где разрешены «сырые» хуки react-redux
	{
		files: ['src/shared/lib/store/**/*.ts'],
		rules: { 'no-restricted-imports': 'off' },
	},

	/* ---------- Тесты ---------- */
	{
		files: ['**/*.test.{ts,tsx}', 'test/**/*.{ts,tsx}'],
		languageOptions: { globals: { ...globals.node } },
		rules: {
			'@typescript-eslint/no-non-null-assertion': 'off',
			'@typescript-eslint/unbound-method': 'off',
		},
	},

	// Конфиги без типизированных правил: eslint.config.js не в tsconfig,
	// у steiger-plugin битый exports (types после import) — TS видит any
	{
		files: ['eslint.config.js', 'steiger.config.ts'],
		...tseslint.configs.disableTypeChecked,
		rules: {
			...tseslint.configs.disableTypeChecked.rules,
			'@typescript-eslint/switch-exhaustiveness-check': 'off',
		},
	},
);
