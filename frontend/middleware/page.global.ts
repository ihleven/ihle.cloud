export default defineNuxtRouteMiddleware((to, from) => {
  const getDepth = (path: string) => {
    return path.split('/').filter(slug => slug.length > 0).length
  }
  const toDepth = getDepth(to.path)
  const fromDepth = getDepth(from.path)

  if (!to.path.startsWith('/kalender')) {
    if (toDepth > fromDepth) {
      to.meta.pageTransition = { name: 'page-left' }
      from.meta.pageTransition = { name: 'page-left' }
    } else {
      to.meta.pageTransition = { name: 'page-right' }
      from.meta.pageTransition = { name: 'page-right' }
    }
  }
  // console.log(
  //   "transition",
  //   fromDepth,
  //   toDepth,
  //   to.meta.pageTransition,
  //   to.path
  // );
})
