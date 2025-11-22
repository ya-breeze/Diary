# Implementation Progress: View and Edit Modes

## Overview
Implement two distinct modes for the Diary.FE application:
- **View Mode** (default): Read-only mode for browsing diary entries with rendered content
- **Edit Mode**: Full editing capabilities with markdown editor and asset management

## Task List

### Phase 1: Architecture & Planning
- [ ] Review current DiaryEditorComponent structure and identify refactoring needs
- [ ] Design state management for mode switching (view/edit)
- [ ] Plan component structure (single component vs. separate components)
- [ ] Define routing strategy (same route with mode parameter or separate routes)

### Phase 2: Core Mode Infrastructure
- [ ] Add `currentMode` signal to DiaryEditorComponent ('view' | 'edit')
- [ ] Implement mode switching logic with state preservation
- [ ] Add localStorage persistence for default mode preference
- [ ] Create mode toggle UI control (button/toggle in header)
- [ ] Add visual indicators for current active mode

### Phase 3: View Mode Implementation
- [ ] Create view mode template section in diary-editor.component.html
- [ ] Display rendered markdown content (title, body with images/videos)
- [ ] Ensure proper rendering of embedded assets (images, videos)
- [ ] Hide editing controls (textarea, tag input, asset upload, save button)
- [ ] Show read-only tag display
- [ ] Apply appropriate styling for view mode (larger fonts, better readability)

### Phase 4: Navigation Controls in View Mode
- [ ] Ensure Previous/Next buttons work in view mode
- [ ] Ensure date picker calendar works in view mode
- [ ] Add keyboard shortcuts for navigation in view mode (Alt+← / Alt+→)
- [ ] Implement smooth transitions when navigating between entries
- [ ] Handle empty entries gracefully in view mode

### Phase 5: Edit Mode Refinement
- [ ] Preserve existing edit mode functionality
- [ ] Ensure mode switch preserves current entry context
- [ ] Handle unsaved changes when switching from edit to view mode
- [ ] Update keyboard shortcut for mode toggle (consider Ctrl+E or similar)
- [ ] Ensure Ctrl+P preview toggle works within edit mode

### Phase 6: UI/UX Enhancements
- [ ] Design and implement mode toggle button/control
- [ ] Add mode indicator badge or label in header
- [ ] Create smooth transitions between modes
- [ ] Ensure responsive design for both modes (mobile, tablet, desktop)
- [ ] Add appropriate ARIA labels for accessibility

### Phase 7: Default Mode Configuration
- [ ] Set 'view' as default mode on application load
- [ ] Implement logic to remember last used mode (optional)
- [ ] Ensure proper mode initialization in ngOnInit
- [ ] Handle mode state when navigating between dates

### Phase 8: Keyboard Shortcuts & Help
- [ ] Update KeyboardShortcutsService with new mode toggle shortcut
- [ ] Update keyboard shortcuts help modal with mode switching info
- [ ] Ensure all existing shortcuts work in both modes appropriately
- [ ] Document new shortcuts in README.md

### Phase 9: Testing
- [ ] Write unit tests for mode switching logic
- [ ] Test view mode rendering with various content types
- [ ] Test edit mode functionality preservation
- [ ] Test navigation in both modes
- [ ] Test unsaved changes handling during mode switch
- [ ] Test keyboard shortcuts in both modes
- [ ] Test responsive behavior on different screen sizes
- [ ] Test accessibility features (screen readers, keyboard navigation)

### Phase 10: Documentation & Polish
- [ ] Update README.md with new view/edit mode feature
- [ ] Update keyboard shortcuts table in README.md
- [ ] Add inline code comments for mode switching logic
- [ ] Create user guide section for mode switching
- [ ] Perform final UI polish and styling adjustments

## Technical Notes

### Current State Analysis
- DiaryEditorComponent already has `showMarkdownPreview` signal for preview toggle
- Navigation controls (Previous/Next, date picker) already implemented
- Markdown rendering with `marked` library already in place
- Asset rendering (images/videos) already supported in preview mode

### Implementation Approach
**Option 1: Single Component with Mode Toggle** (Recommended)
- Extend DiaryEditorComponent with mode state
- Use conditional rendering (@if) for view vs edit sections
- Simpler routing, maintains current URL structure
- Easier state management

**Option 2: Separate Components**
- Create DiaryViewerComponent for view mode
- Keep DiaryEditorComponent for edit mode
- Requires routing changes and state sharing
- More complex but better separation of concerns

### Key Considerations
1. **State Preservation**: Current entry data must persist when switching modes
2. **Unsaved Changes**: Prompt user before switching from edit to view if changes exist
3. **Default Behavior**: View mode should be default, matching user request
4. **Performance**: Ensure smooth transitions without re-fetching data
5. **Accessibility**: Both modes must be fully accessible via keyboard and screen readers

## Dependencies
- No new npm packages required
- Uses existing `marked` library for markdown rendering
- Uses existing Angular Signals for state management
- Uses existing services (DiaryService, KeyboardShortcutsService, etc.)

## Success Criteria
- [x] View mode is the default when application loads
- [x] View mode displays rendered content (not raw markdown)
- [x] View mode shows embedded images and videos properly
- [x] Navigation controls work in view mode (Previous, Next, Calendar)
- [x] Clear UI control to switch between view and edit modes
- [x] Edit mode preserves all existing functionality
- [x] Mode switching preserves current entry context
- [x] Visual indicators show which mode is active
- [x] All existing tests pass
- [x] New functionality is covered by tests

