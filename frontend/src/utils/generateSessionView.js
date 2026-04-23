export function buildGenerateSessionTimeline(currentResults, currentSessionLabel) {
  if (!Array.isArray(currentResults) || currentResults.length === 0) {
    return []
  }

  return [{ label: currentSessionLabel, items: currentResults }]
}

export function clearGenerateSessionResults() {
  return []
}

export function hasGenerateSessionContent(currentResults) {
  return Array.isArray(currentResults) && currentResults.length > 0
}
