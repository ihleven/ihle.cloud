export const useAuth = () => {
  const auth = useState<{}>('useauth', () => ({}))

  function login(target: string) {
    const redirect = typeof target === 'string' && target !== '' ? `?redirect=${target}` : ''
    navigateTo('/api/auth/login' + redirect, { external: true })
  }

  function logout() {
    $fetch('/api/auth/logout')
    auth.value = {}
    navigateTo('/welcome')
  }

  return {
    auth,
    logout,
    login,
  }
}
