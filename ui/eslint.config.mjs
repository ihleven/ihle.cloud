// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'
import eslintPluginBetterTailwindcss from 'eslint-plugin-better-tailwindcss'

export default withNuxt({
  plugins: {
    'better-tailwindcss': eslintPluginBetterTailwindcss,
  },
  settings: {
    'better-tailwindcss': {
      entryPoint: 'app/assets/css/main.css',
    },
  },
  rules: {
    ...eslintPluginBetterTailwindcss.configs.recommended.rules,
    'better-tailwindcss/enforce-consistent-line-wrapping': 'off',
    // The geheimtipp layer carries a few classes of its own — named grid areas,
    // the pool's purple, outlined text — for things Tailwind has no utility for.
    // They are prefixed so they can be recognised by shape rather than listed.
    //
    // The front page's tiles are the same arrangement: a photograph that grows
    // under the pointer and drifts as the page scrolls is three transforms on
    // nested elements, which is a stylesheet's job and not a row of utilities.
    // The ones applied through the module list were never flagged — the linter
    // cannot see a bound class — so this only makes the rule agree with itself.
    'better-tailwindcss/no-unregistered-classes': ['error', { ignore: ['^ght-', '^tile(-|$)'] }],
    //
    'no-console': 'off',
    'vue/multi-word-component-names': 'off',
    'vue/max-attributes-per-line': ['error', {
      singleline: {
        max: 5,
      },
      multiline: {
        max: 3,
      },
    }],
    'vue/singleline-html-element-content-newline': 'off',
    'vue/multiline-html-element-content-newline': 'off',
  },
})
