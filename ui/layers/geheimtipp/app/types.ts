// What the pool's backend actually returns.
//
// Written down because the payloads carry their own conventions and are not
// consistent between endpoints: the edition list calls a thing `name` and
// `Klassements`, the edition detail calls the same things `label` and
// `klassements`. Two types rather than one type with everything optional, so a
// page cannot quietly read a field the endpoint it used does not send.
//
// Field names and their types come from the live payloads, not from the pool's
// own frontend. Only the fields something here renders are named; the rest of
// each payload is scoring bookkeeping that no screen shows.

/** GET /aktuell — who is signed in, and which rounds to offer. */
export type Aktuell = {
  /** The signed-in login, or "" — this is the session probe. */
  authkey: string
  /** A string, not a number: "2027". */
  edition: string
  geheimtipper: Geheimtipper | null
  /** The signed-in person's entry in the current edition, or null. */
  registration: EditionTipper | null
  spieltage: {
    last: RoundWindow | null
    current: RoundWindow | null
    next: RoundWindow | null
  } | null
}

export type RoundWindow = {
  round: number
  from: string
  to: string
}

/** POST /login — form-encoded `username` + `password`. */
export type LoginResult = {
  Username: string
  /** The backend sets no cookie; whoever calls this has to. */
  JWT: string
}

/** An edition as the list returns it: GET /ausgaben */
export type AusgabeSummary = {
  id: number
  code: string
  name: string
  season: number
  competitions: string[]
  /** True for every edition ever run. Not the current one — /aktuell knows that. */
  aktuell: boolean
  numTipper: number
  winner: string[] | null

  // Sent as null by the list and populated by the detail, under different
  // names. Left unknown here because nothing should read them from a summary.
  wettbewerbe: unknown
  Klassements: unknown
  tipper: unknown
  tipperMap: unknown
}

/** An edition as the detail returns it: GET /ausgaben/:code */
export type AusgabeDetail = {
  id: number
  code: string
  /** The list calls this `name`. */
  label: string
  season: number
  competitions: string[]
  wettbewerbe: Wettbewerb[]
  tipper: EditionTipper[]
  /** The list calls this `Klassements`, capitalised, and sends null. */
  klassements: Klassement[] | null
}

/** A competition within an edition. */
export type Wettbewerb = {
  edt: number
  number: number
  comp: string
  season: number
  code: string
  numTeilnehmer: number
  numTippps: number
  winner: string[]
}

/** One of an edition's rankings — a jersey, and the rule behind it. */
export type Klassement = {
  /** The colour key, matching the keys in an eval's `trikots`. */
  trikot: string
  rolle: string
  isGesamt: boolean
  isMain: boolean
  isOfficial: boolean
  Winner: string[]
  Punkte: number
  Description: string
}

/**
 * A tipper's standing within an edition. The ranking is built from this, which
 * is why it arrives with the edition rather than from a call of its own.
 */
export type EditionTipper = {
  id: number
  edt: number
  ght: string
  name: string
  motto: string
  /** An id, not a filename: the image is /media/avatars/<avatar>/<size>. */
  avatar: number
  registriert: string
  platz: number
  tipps: number
  runden: number
  pkt: number
  tendenz: number
  treffer: number
  /** Keyed by competition code. Empty on the edition list, filled for yourself. */
  teilnahmen: Record<string, Teilnahme>
}

/** How one tipper did in one competition of an edition. */
export type Teilnahme = {
  login: string
  runden: number
  tipps: number
  platz: number
  punkte: number
}

/** A person, as GET /geheimtipper returns them. */
export type Geheimtipper = {
  login: string
  vorname: string
  nachname: string
  /**
   * Present, and only ever shown to whoever it belongs to.
   *
   * GET /aktuell returns the signed-in person's own record, which is where the
   * account menu takes it from. GET /geheimtipper returns one of these for
   * everybody who has ever played, addresses included — nothing here reads that
   * endpoint's copy, and nothing should.
   */
  email: string
  verein?: string
  teams: string[]
  seit: string
  turniernr: number
  avatar: number | null
}

/** One person, as GET /geheimtipper/:login returns them. */
export type GeheimtipperDetail = Geheimtipper & {
  /** The editions they have played, with where they finished. */
  Registrierungen: EditionTipper[]
}

/**
 * A matchday: GET /comprounds/:comp/:season/:round
 *
 * One request carries the whole matrix screen — the rows, the columns, both
 * footer rows and the league table beneath it.
 */
export type CompRound = {
  comp: string
  season: number
  round: number
  from: string
  to: string
  stats: RoundStats
  matches: Match[]
  tabelle: TableRow[]
  /** Keyed by login. The matrix takes its column list from here. */
  evals: Record<string, Eval>
  /** Keyed by team code. */
  teams: Record<string, Team>
  compSeason: CompSeason
}

export type RoundStats = {
  ntipps: number
  ntipper: number
  pktMin: number
  pktAvg: number
  pktMax: number
}

/**
 * A single match: GET /matches/:id.
 *
 * The same shape a round carries, plus the round itself — which the list form
 * leaves null, because there it would only repeat what was asked for.
 */
export type MatchDetail = Match & {
  round: MatchRound
}

export type MatchRound = {
  comp: string
  season: number
  num: number
  /** What the round calls itself: "3". */
  code: string
  from: string
  to: string
  numMatches: number
}

/**
 * What PUT /comprounds/:comp/:season/:round/tipps answers.
 *
 * One entry per match that was sent, keyed by match id. A refusal — past the
 * kick-off, not registered — arrives as an `Error` on that entry rather than a
 * failed request, so some tips of a batch can be taken and others declined.
 */
export type TippResponse = {
  Round: CompRound
  Status: Record<string, TippStatus>
}

export type TippStatus = {
  Error: { cause: string } | null
  Odds: { tipps: Record<string, Tipp> | null } | null
}

export type Match = {
  /** Season, round and number run together: 20270301. */
  id: number
  comp: string
  season: number
  stage: string
  no: number
  group: string
  kickoff: string
  /** Home first, away second — both are team codes. */
  teams: [string, string]
  /** Already formatted by the backend: "3:1", or "-:-" before it is played. */
  result: string
  /** Every tipper's tip, keyed by login. Returned without a token. */
  odds: { tipps: Record<string, Tipp> | null } | null

  // `round` is present but null inside a round; GET /matches/:id fills it in.
  // See MatchDetail.
}

export type Tipp = {
  login: string
  /** A pair, unlike the match's own formatted result. Absent when not tipped. */
  Result: [number, number] | null
  /** 2 exact, 1 tendency, 0 wrong — what the badge on the cell shows. */
  Punkte: number
  toto: number
  placed: string
}

/** One row of the league table. */
export type TableRow = {
  team: string
  POS: number
  /** Position delta against the previous matchday: the movement arrow. */
  PD: number
  PLD: number
  W: number
  D: number
  L: number
  GF: number
  GA: number
  GD: number
  /** Goal quotient — shown instead of GD when the season says `quot`. */
  GQ: number
  PTS: number
  /** Two-point counting — shown instead of PTS when the season says `pkt2`. */
  PP: number
  MP: number
}

/** How one tipper did on one matchday, plus where they stand overall. */
export type Eval = {
  /** The day's result: the Tageswertung. */
  spieltag: Standing | null
  /** Overall standings, keyed by jersey colour. Only GELB is filled in so far. */
  trikots: Record<string, Trikot>
  ntipps: number
  runden: number
  ntendenz: number
  ntreffer: number
}

export type Standing = {
  platz: number
  punkte: number
  plzpkt: number
}

export type Trikot = {
  trikot: string
  /** Where they stand overall — what the matrix sorts its columns by. */
  platz: number
  punkte: number
  platzDiff: number
  /** Their position on this matchday alone. */
  pos: number
  rndPts: number
}

export type Team = {
  code: string
  name: string
  /** Always "". Crests are files we ship, not data. */
  logo: string
}

/** A season: GET /compseasons/:comp/:season */
export type CompSeason = {
  season: number
  nteams: number
  nrounds: number
  nmatches: number
  /** Count goals as a quotient rather than a difference. */
  quot: boolean
  /** Two-point counting rather than three. */
  pkt2: boolean
  /** Null inside a CompRound; the list of rounds only comes from /compseasons. */
  rounds: CompSeasonRound[] | null
}

export type CompSeasonRound = {
  num: number
  phase: string
  from: string
  to: string
  code: string
}

/**
 * An uploaded image: GET /avatars.
 *
 * `id` is how it is addressed — /media/avatars/<id>/<size>. `name` is the
 * original filename and the backend does not route by it.
 */
export type AvatarFile = {
  id: number
  /** The login it belongs to, or "" for the unclaimed ones. */
  ght: string
  /** The original filename. */
  name: string
  /** Where it sits in the pool's own store. */
  file: string
  created: string
  size: number
  w: number
  h: number
  format: string
  mimetype: string
  /** Colour mode as the pool's image library reports it: "RGB", "P", … */
  mode: string
  labels: string[]
  /** A loose grouping the pool gave a batch of pictures, e.g. "star-wars". */
  topic: string
  public: boolean
  active: boolean
  account_id: number
}
