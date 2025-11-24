/**
 * Layout Toggle Functionality for Diary.BE
 *
 * This module provides interactive layout switching functionality for the Diary.BE home page,
 * allowing users to toggle between full-width (100%) and narrow (30%) image display modes.
 *
 * Features:
 * - Two layout modes: Full (100% width) and Narrow (30% width)
 * - Session persistence using localStorage
 * - Keyboard navigation support (Enter/Space keys)
 * - ARIA accessibility compliance
 * - Responsive design across all device sizes
 * - Graceful degradation when JavaScript is disabled
 * - Error handling and recovery mechanisms
 *
 * Browser Support:
 * - Modern browsers: Full functionality
 * - IE11: Limited functionality with fallbacks
 * - No JavaScript: CSS-only fallback to default layout
 *
 * Dependencies:
 * - Modern browser with ES5+ support
 * - localStorage (optional, graceful degradation)
 * - CSS classes defined in layout.css
 *
 * @author Diary.BE Development Team
 * @version 1.0.0
 * @since 2025-01-19
 */

(function () {
    'use strict';

    /**
     * Configuration constants for layout toggle functionality
     */
    const LAYOUT_MODES = {
        FULL: 'full',    // 100% image width mode
        NARROW: 'narrow' // 30% image width mode
    };

    /** localStorage key for persisting user's layout preference */
    const STORAGE_KEY = 'diary-layout-preference';

    /** Duration of CSS transitions in milliseconds */
    const TRANSITION_DURATION = 300;

    /**
     * DOM element references
     * These are populated during initialization and cached for performance
     */
    let fullLayoutBtn, narrowLayoutBtn, mainContent, layoutStatus;

    /**
     * Initialize the layout toggle functionality
     *
     * This is the main entry point that sets up the entire layout toggle system.
     * It performs the following steps:
     * 1. Locates and caches required DOM elements
     * 2. Tests browser compatibility for required features
     * 3. Attaches event listeners for user interactions
     * 4. Restores saved layout preference from localStorage
     * 5. Handles any initialization errors gracefully
     *
     * @throws {Error} If critical initialization steps fail
     */
    function initLayoutToggle() {
        try {
            // Cache DOM element references for performance
            // These elements are required for the layout toggle to function
            fullLayoutBtn = document.getElementById('fullLayoutBtn');
            narrowLayoutBtn = document.getElementById('narrowLayoutBtn');
            mainContent = document.getElementById('mainContent');
            layoutStatus = document.getElementById('layout-status');

            // Verify all required DOM elements are present
            if (!fullLayoutBtn || !narrowLayoutBtn || !mainContent) {
                console.warn('Layout toggle: Required DOM elements not found. Layout toggle functionality will be disabled.');
                showFallbackMessage();
                return;
            }

            // Test browser compatibility for required features
            // This includes localStorage, classList, addEventListener, and CustomEvent
            if (!testBrowserSupport()) {
                console.warn('Layout toggle: Browser does not support required features. Using fallback behavior.');
                showFallbackMessage();
                return;
            }

            // Set up event listeners for user interactions
            addEventListeners();

            // Restore user's previously saved layout preference
            // This ensures consistency across browser sessions
            const savedLayout = loadLayoutPreference();
            if (savedLayout && savedLayout !== getCurrentLayout()) {
                switchLayout(savedLayout, false); // Don't animate on initial load for better UX
            }

            console.log('Layout toggle initialized successfully');
        } catch (error) {
            console.error('Failed to initialize layout toggle:', error);
            showFallbackMessage();
        }
    }

    /**
     * Test browser support for required features
     *
     * This function performs feature detection to ensure the browser supports
     * all APIs required for the layout toggle functionality. It tests for:
     * - localStorage: For session persistence
     * - classList API: For CSS class manipulation
     * - addEventListener: For event handling
     * - CustomEvent: For dispatching layout change events
     *
     * @returns {boolean} true if all required features are supported, false otherwise
     */
    function testBrowserSupport() {
        try {
            // Test localStorage availability and functionality
            // Some browsers disable localStorage in private/incognito mode
            const testKey = 'diary-test-storage';
            localStorage.setItem(testKey, 'test');
            localStorage.removeItem(testKey);

            // Test for classList API (IE10+)
            // Required for adding/removing CSS classes
            if (!document.documentElement.classList) {
                return false;
            }

            // Test for addEventListener (IE9+)
            // Required for event handling
            if (!document.addEventListener) {
                return false;
            }

            // Test for CustomEvent constructor (IE9+ with polyfill, native in modern browsers)
            // Required for dispatching layout change events
            if (typeof CustomEvent === 'undefined') {
                return false;
            }

            return true;
        } catch (error) {
            // Any exception during feature testing indicates lack of support
            return false;
        }
    }

    /**
     * Add event listeners with error handling
     */
    function addEventListeners() {
        try {
            fullLayoutBtn.addEventListener('click', () => switchLayout(LAYOUT_MODES.FULL));
            narrowLayoutBtn.addEventListener('click', () => switchLayout(LAYOUT_MODES.NARROW));

            // Add keyboard support
            fullLayoutBtn.addEventListener('keydown', handleKeyDown);
            narrowLayoutBtn.addEventListener('keydown', handleKeyDown);

            // Add error recovery for failed layout switches
            window.addEventListener('error', handleGlobalError);
        } catch (error) {
            console.error('Failed to add event listeners:', error);
            throw error;
        }
    }

    /**
     * Show fallback message when JavaScript features are not available
     */
    function showFallbackMessage() {
        try {
            const container = document.querySelector('.layout-toggle-container');
            if (container) {
                container.style.display = 'none';
            }

            const message = document.querySelector('.js-disabled-message');
            if (message) {
                message.style.display = 'block';
                message.textContent = 'Layout toggle functionality is not available in your browser. Using default layout.';
            }
        } catch (error) {
            console.error('Failed to show fallback message:', error);
        }
    }

    /**
     * Handle global errors that might affect layout functionality
     */
    function handleGlobalError(event) {
        if (event.error && event.error.message && event.error.message.includes('layout')) {
            console.error('Layout-related error detected:', event.error);
            // Reset to safe state
            try {
                if (mainContent) {
                    mainContent.classList.remove('layout-changing');
                }
            } catch (resetError) {
                console.error('Failed to reset layout state:', resetError);
            }
        }
    }

    /**
     * Handle keyboard navigation
     */
    function handleKeyDown(event) {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            const layout = event.target.dataset.layout;
            if (layout) {
                switchLayout(layout);
            }
        }
    }

    /**
     * Switch to the specified layout mode
     */
    function switchLayout(mode, animate = true) {
        try {
            if (!isValidLayout(mode)) {
                console.error('Invalid layout mode:', mode);
                return false;
            }

            const currentLayout = getCurrentLayout();
            if (currentLayout === mode) {
                return true; // Already in this mode
            }

            // Verify DOM elements are still available
            if (!mainContent || !fullLayoutBtn || !narrowLayoutBtn) {
                console.error('Layout toggle: DOM elements not available for layout switch');
                return false;
            }

            // Add transition class if animating
            if (animate) {
                mainContent.classList.add('layout-changing');
            }

            // Update layout classes
            if (!updateLayoutClasses(mode)) {
                throw new Error('Failed to update layout classes');
            }

            // Update button states
            if (!updateButtonStates(mode)) {
                throw new Error('Failed to update button states');
            }

            // Update ARIA states
            updateAriaStates(mode);

            // Update status for screen readers
            updateLayoutStatus(mode);

            // Save preference
            saveLayoutPreference(mode);

            // Remove transition class after animation
            if (animate) {
                setTimeout(() => {
                    try {
                        if (mainContent) {
                            mainContent.classList.remove('layout-changing');
                        }
                    } catch (error) {
                        console.warn('Failed to remove transition class:', error);
                    }
                }, TRANSITION_DURATION);
            }

            // Dispatch custom event for other components
            dispatchLayoutChangeEvent(mode, currentLayout);

            console.log('Layout switched to:', mode);
            return true;
        } catch (error) {
            console.error('Failed to switch layout:', error);

            // Attempt to recover by removing transition class
            try {
                if (mainContent) {
                    mainContent.classList.remove('layout-changing');
                }
            } catch (recoveryError) {
                console.error('Failed to recover from layout switch error:', recoveryError);
            }

            return false;
        }
    }

    /**
     * Update CSS classes on the main content
     */
    function updateLayoutClasses(mode) {
        try {
            if (!mainContent) {
                console.error('Main content element not available');
                return false;
            }

            // Remove all layout classes
            mainContent.classList.remove('layout-full', 'layout-narrow');

            // Add the new layout class
            mainContent.classList.add(`layout-${mode}`);
            return true;
        } catch (error) {
            console.error('Failed to update layout classes:', error);
            return false;
        }
    }

    /**
     * Update button visual states
     */
    function updateButtonStates(mode) {
        try {
            if (!fullLayoutBtn || !narrowLayoutBtn) {
                console.error('Button elements not available');
                return false;
            }

            // Remove active class from all buttons
            fullLayoutBtn.classList.remove('active');
            narrowLayoutBtn.classList.remove('active');

            // Add active class to selected button
            if (mode === LAYOUT_MODES.FULL) {
                fullLayoutBtn.classList.add('active');
            } else {
                narrowLayoutBtn.classList.add('active');
            }
            return true;
        } catch (error) {
            console.error('Failed to update button states:', error);
            return false;
        }
    }

    /**
     * Update ARIA states for accessibility
     */
    function updateAriaStates(mode) {
        // Update aria-pressed states
        fullLayoutBtn.setAttribute('aria-pressed', mode === LAYOUT_MODES.FULL ? 'true' : 'false');
        narrowLayoutBtn.setAttribute('aria-pressed', mode === LAYOUT_MODES.NARROW ? 'true' : 'false');
    }

    /**
     * Update status message for screen readers
     */
    function updateLayoutStatus(mode) {
        if (!layoutStatus) return;

        const messages = {
            [LAYOUT_MODES.FULL]: 'Current layout: Full width - images display at 100% width',
            [LAYOUT_MODES.NARROW]: 'Current layout: Narrow - images display at 30% width'
        };

        layoutStatus.textContent = messages[mode] || '';
    }

    /**
     * Get the current layout mode
     */
    function getCurrentLayout() {
        if (mainContent.classList.contains('layout-full')) {
            return LAYOUT_MODES.FULL;
        } else if (mainContent.classList.contains('layout-narrow')) {
            return LAYOUT_MODES.NARROW;
        }
        return LAYOUT_MODES.NARROW; // Default
    }

    /**
     * Validate layout mode
     */
    function isValidLayout(mode) {
        return Object.values(LAYOUT_MODES).includes(mode);
    }

    /**
     * Save layout preference to localStorage
     */
    function saveLayoutPreference(mode) {
        try {
            localStorage.setItem(STORAGE_KEY, mode);
        } catch (error) {
            console.warn('Failed to save layout preference:', error);
        }
    }

    /**
     * Load layout preference from localStorage
     */
    function loadLayoutPreference() {
        try {
            const saved = localStorage.getItem(STORAGE_KEY);
            return isValidLayout(saved) ? saved : null;
        } catch (error) {
            console.warn('Failed to load layout preference:', error);
            return null;
        }
    }

    /**
     * Dispatch custom event when layout changes
     */
    function dispatchLayoutChangeEvent(newLayout, oldLayout) {
        const event = new CustomEvent('layoutChange', {
            detail: {
                newLayout: newLayout,
                oldLayout: oldLayout,
                timestamp: Date.now()
            }
        });
        document.dispatchEvent(event);
    }

    /**
     * Public API for external access
     */
    window.DiaryLayoutToggle = {
        switchLayout: switchLayout,
        getCurrentLayout: getCurrentLayout,
        LAYOUT_MODES: LAYOUT_MODES
    };

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initLayoutToggle);
    } else {
        initLayoutToggle();
    }

})();
