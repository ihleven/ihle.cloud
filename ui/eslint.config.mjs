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
