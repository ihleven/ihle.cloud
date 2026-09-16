import { describe, expect, it } from 'vitest'
import {
  agreed,
  albumMatches,
  albumTitle,
  albumYear,
  isAudio,
  isCover,
  leafDirs,
  parentDir,
  sortAlbums,
  sortTracks,
  trackMeta,
  trackTitle,
} from '../app/utils/musik'

describe('albumTitle', () => {
  it('drops the year, which is shown on its own', () => {
    expect(albumTitle('1987 - Appetite For Destruction (lameV3A)')).toBe('Appetite For Destruction')
  })

  it('drops the encoder, which is a fact about the file', () => {
    expect(albumTitle('1991 - Use Your Illusion I (lameV3A)')).toBe('Use Your Illusion I')
    expect(albumTitle('1987 - Raise Your Fist And Yell (lameV3)')).toBe('Raise Your Fist And Yell')
  })

  it('leaves a folder that carries neither', () => {
    expect(albumTitle('Trash')).toBe('Trash')
    expect(albumTitle('Hey Stoopid')).toBe('Hey Stoopid')
  })

  // The reason the encoder is matched rather than "the last bracket": a title
  // may end in one, and dropping it would rename the album.
  it('keeps a parenthesis that belongs to the title', () => {
    expect(albumTitle('1978 - Live At Budokan (Remastered)')).toBe('Live At Budokan (Remastered)')
    expect(albumTitle('Alive II (Live)')).toBe('Alive II (Live)')
  })

  it('reads the folder and not the path above it', () => {
    expect(albumTitle('Alice Cooper/1989 - Trash (lameV3)')).toBe('Trash')
  })

  it('never returns nothing, whatever the folder is called', () => {
    expect(albumTitle('1987 - (lameV3A)')).toBe('1987 - (lameV3A)')
  })
})

describe('albumYear', () => {
  it('reads a leading year', () => {
    expect(albumYear('1987 - Appetite For Destruction (lameV3A)')).toBe('1987')
  })

  it('is absent where the folder carries none', () => {
    expect(albumYear('Trash')).toBeUndefined()
  })

  // A number that is not a year, and a year that is not at the front: neither
  // is the album's date.
  it('does not take any four digits it finds', () => {
    expect(albumYear('1234 - Nope')).toBeUndefined()
    expect(albumYear('Live 1987')).toBeUndefined()
  })
})

describe('trackTitle', () => {
  it('splits the position from the title', () => {
    expect(trackTitle('01. Hey Stoopid.mp3')).toEqual({ number: 1, title: 'Hey Stoopid' })
    expect(trackTitle('11. Dirty Dreams.mp3')).toEqual({ number: 11, title: 'Dirty Dreams' })
  })

  it('reads the separators the shelf actually uses', () => {
    expect(trackTitle('02 - Spark In The Dark.mp3')).toEqual({ number: 2, title: 'Spark In The Dark' })
    expect(trackTitle('03 House Of Fire.mp3')).toEqual({ number: 3, title: 'House Of Fire' })
    expect(trackTitle('04_Why Trust You.mp3')).toEqual({ number: 4, title: 'Why Trust You' })
  })

  it('keeps an apostrophe and the rest of the name intact', () => {
    expect(trackTitle('05. Only My Heart Talkin\'.mp3')).toEqual({ number: 5, title: 'Only My Heart Talkin\'' })
  })

  // Removing the number here would leave the track with no name at all.
  it('leaves a title that is only a number', () => {
    expect(trackTitle('1984.mp3')).toEqual({ title: '1984' })
  })

  it('leaves an unnumbered track unnumbered rather than guessing', () => {
    expect(trackTitle('Hey Stoopid.mp3')).toEqual({ title: 'Hey Stoopid' })
  })
})

describe('sortTracks', () => {
  it('puts an album in the order it plays, not the order it was listed', () => {
    const tracks = [
      { name: '10. J.mp3', path: 'a/10. J.mp3', title: 'J', number: 10 },
      { name: '02. B.mp3', path: 'a/02. B.mp3', title: 'B', number: 2 },
      { name: '01. A.mp3', path: 'a/01. A.mp3', title: 'A', number: 1 },
    ]
    expect(sortTracks(tracks).map(t => t.number)).toEqual([1, 2, 10])
  })

  it('leaves the unnumbered at the end rather than first', () => {
    const tracks = [
      { name: 'Bonus.mp3', path: 'a/Bonus.mp3', title: 'Bonus' },
      { name: '01. A.mp3', path: 'a/01. A.mp3', title: 'A', number: 1 },
    ]
    expect(sortTracks(tracks).map(t => t.title)).toEqual(['A', 'Bonus'])
  })
})

describe('sortAlbums', () => {
  it('reads newest first, and puts the undated together at the end', () => {
    const albums = [
      { path: 'b', title: 'B', year: '1987', tracks: [] },
      { path: 'c', title: 'C', tracks: [] },
      { path: 'a', title: 'A', year: '1991', tracks: [] },
    ]
    expect(sortAlbums(albums).map(a => a.title)).toEqual(['A', 'B', 'C'])
  })
})

// The rule both readings of the shelf share. If they disagreed here they would
// return different albums and the comparison between them would mean nothing.
describe('leafDirs', () => {
  it('takes the folders that hold no other folder', () => {
    expect(leafDirs(['Trash', 'Hey Stoopid'])).toEqual(['Trash', 'Hey Stoopid'])
  })

  it('drops a folder that only holds other folders', () => {
    const dirs = ['Alice Cooper', 'Alice Cooper/1989 - Trash', 'Alice Cooper/1991 - Hey Stoopid']
    expect(leafDirs(dirs)).toEqual(['Alice Cooper/1989 - Trash', 'Alice Cooper/1991 - Hey Stoopid'])
  })

  // A prefix is not a parent: "Trash" does not contain "Trashcan".
  it('does not mistake a shared prefix for containment', () => {
    expect(leafDirs(['Trash', 'Trashcan'])).toEqual(['Trash', 'Trashcan'])
  })

  it('copes with a shelf of mixed depth', () => {
    const dirs = ['Trash', 'Alice Cooper', 'Alice Cooper/1989 - Hey Stoopid']
    expect(leafDirs(dirs)).toEqual(['Trash', 'Alice Cooper/1989 - Hey Stoopid'])
  })
})

describe('parentDir', () => {
  it('is the folder a file sits in', () => {
    expect(parentDir('Trash/01. Poison.mp3')).toBe('Trash')
    expect(parentDir('Alice Cooper/Trash/01. Poison.mp3')).toBe('Alice Cooper/Trash')
  })

  it('is the shelf itself for something at the top', () => {
    expect(parentDir('cover.jpeg')).toBe('')
  })
})

describe('isAudio and isCover', () => {
  it('knows what can be played', () => {
    expect(isAudio('01. Poison.mp3')).toBe(true)
    expect(isAudio('01. Poison.flac')).toBe(true)
    expect(isAudio('Alice Cooper - Trash.log')).toBe(false)
    expect(isAudio('cover.jpeg')).toBe(false)
  })

  it('knows the sleeve from any other picture', () => {
    expect(isCover('cover.jpeg')).toBe(true)
    expect(isCover('folder.jpg')).toBe(true)
    expect(isCover('back.jpg')).toBe(false)
  })
})

describe('albumMatches', () => {
  const album = {
    path: '1987 - Appetite For Destruction (lameV3A)',
    title: 'Appetite For Destruction',
    year: '1987',
    tracks: [{ name: '01. Welcome To The Jungle.mp3', path: 'x', title: 'Welcome To The Jungle', number: 1 }],
  }

  it('matches nothing typed by matching everything', () => {
    expect(albumMatches(album, '')).toBe(true)
  })

  it('matches the title however it is cased', () => {
    expect(albumMatches(album, 'appetite')).toBe(true)
    expect(albumMatches(album, 'DESTRUCTION')).toBe(true)
  })

  it('matches the year, which is how somebody looks for a period', () => {
    expect(albumMatches(album, '1987')).toBe(true)
  })

  it('matches a track, so an album is findable by what is on it', () => {
    expect(albumMatches(album, 'jungle')).toBe(true)
  })

  // Somebody looking for Motörhead types Motorhead.
  it('ignores accents in both directions', () => {
    const umlaut = { path: 'x', title: 'Motörhead', tracks: [] }
    expect(albumMatches(umlaut, 'motorhead')).toBe(true)
    expect(albumMatches(umlaut, 'motörhead')).toBe(true)
  })

  it('does not match what is not there', () => {
    expect(albumMatches(album, 'nirvana')).toBe(false)
  })
})

describe('agreed', () => {
  it('is the value where they all say the same', () => {
    expect(agreed(['Alice Cooper', 'Alice Cooper', 'Alice Cooper'])).toBe('Alice Cooper')
  })

  // Nothing rather than the commonest: where the tracks disagree there is no
  // album-level answer, and taking the first would silence the rows that differ.
  it('is nothing where they disagree', () => {
    expect(agreed(['Alice Cooper', 'Guns N\' Roses'])).toBe('')
  })

  it('ignores the ones that say nothing at all', () => {
    expect(agreed(['Alice Cooper', undefined, ''])).toBe('Alice Cooper')
    expect(agreed([undefined, undefined])).toBe('')
  })
})

describe('trackMeta', () => {
  const heading = { artist: 'Alice Cooper', album: 'Trash', year: '1989', genre: 'Hard Rock' }

  // The whole point: an album whose tracks all agree with its heading has
  // nothing further to say, and said it ten times before this.
  it('says nothing where the track agrees with the heading', () => {
    const tag = { artist: 'Alice Cooper', album: 'Trash', year: 1989, genre: 'Hard Rock' }
    expect(trackMeta(tag, heading)).toBe('')
  })

  it('names the guest on one song', () => {
    const tag = { artist: 'Alice Cooper & Ozzy', album: 'Trash', year: 1989, genre: 'Hard Rock' }
    expect(trackMeta(tag, heading)).toBe('Alice Cooper & Ozzy')
  })

  // A compilation: the heading can name no single artist, so every track names
  // its own.
  it('names every artist where the heading can name none', () => {
    const various = { artist: '', album: 'Sampler', year: '1994', genre: '' }
    expect(trackMeta({ artist: 'Nirvana', album: 'Sampler', year: 1994 }, various)).toBe('Nirvana')
    expect(trackMeta({ artist: 'Hole', album: 'Sampler', year: 1994 }, various)).toBe('Hole')
  })

  it('names a year the rest of the album does not share', () => {
    const tag = { artist: 'Alice Cooper', album: 'Trash', year: 1991, genre: 'Hard Rock' }
    expect(trackMeta(tag, heading)).toBe('1991')
  })

  it('names a second album inside one folder', () => {
    const tag = { artist: 'Alice Cooper', album: 'Hey Stoopid', year: 1989, genre: 'Hard Rock' }
    expect(trackMeta(tag, heading)).toBe('Hey Stoopid')
  })

  it('names several at once, in the order the heading reads', () => {
    const tag = { artist: 'Ozzy', album: 'Hey Stoopid', year: 1991, genre: 'Metal' }
    expect(trackMeta(tag, heading)).toBe('Ozzy · Hey Stoopid · 1991 · Metal')
  })

  // An empty field is not a disagreement: a track that simply carries no genre
  // is not saying the album's genre is wrong.
  it('treats a missing field as nothing to report', () => {
    const tag = { artist: 'Alice Cooper', album: 'Trash', year: 1989 }
    expect(trackMeta(tag, heading)).toBe('')
  })
})
