## Why

The EntryEditor currently auto-fetches tag suggestions 4 seconds after the user stops typing. This fires on mount too (when `aiEnabled` flips from false to true), so suggestions appear without the user touching anything. The user finds this distracting and wants full control over when suggestions are requested.

## What Changes

- Remove the debounced `useEffect` that auto-triggers `fetchSuggestions()` in `EntryEditor`
- Remove the unused `SUGGEST_DEBOUNCE_MS` constant
- The explicit "Suggest Tags" (✨) button remains as the only way to fetch suggestions
- Pre-loaded `pendingTags` from the backend still surface as chips on editor open (unchanged)

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `ai-tagging`: Remove the "In-editor debounced auto-suggest" requirement; the explicit button becomes the sole attended trigger for suggestion fetching

## Impact

- `next-frontend/src/components/diary/EntryEditor.tsx`: remove ~8 lines (constant + useEffect)
- `openspec/specs/ai-tagging/spec.md`: remove the "In-editor debounced auto-suggest" requirement block and update the trigger enumeration on lines 11 and 53
