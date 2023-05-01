module.exports = {
  printWidth: 120,
  tabWidth: 2,
  semi: false,
  singleQuote: true,
  bracketSpacing: true,
  bracketSameLine: false,
  arrowParens: 'avoid',
  vueIndentScriptAndStyle: true,

  plugins: [require('prettier-plugin-tailwindcss')],
  tailwindConfig: './frontend/tailwind.config.js',
}
