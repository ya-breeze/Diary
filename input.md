Analyze the OpenAPI specification file located at `/home/ek/work/Diary.FE/api/openapi.yaml` to understand the backend API structure for the personal diary application.

Based on this analysis, create a comprehensive implementation plan and save it to a new file called `IMPLEMENTATION.md` in the workspace root. The plan should include:

1. **API Analysis Summary**: Document the key endpoints, data models, authentication requirements, and API capabilities discovered from the OpenAPI spec
2. **Architecture Design**: Propose the Angular application structure including:
   - Module organization (feature modules, shared modules, core module)
   - Component hierarchy and routing structure
   - Service layer design for API integration
   - State management approach (if needed)
   - Authentication/authorization implementation
3. **Implementation Phases**: Break down the development into logical phases with clear milestones
4. **Detailed Task List**: Create a structured, prioritized task list covering:
   - Project initialization and setup
   - Core infrastructure (routing, HTTP interceptors, error handling)
   - Feature implementation (one section per major feature/endpoint group)
   - UI/UX components
   - Testing strategy
   - Deployment considerations

Also create a corresponding task list using the task management tools to track the implementation progress.

The plan should be detailed enough to serve as a complete development roadmap for building the Angular frontend application.

Take into account:

- Create Makefile for common tasks
- API supports multiple users
- API uses "items" to represent diary entries
- API supports synchronization of changes, but it's needed for mobile app only. Ignore it for now.
- API uses "assets" to represent media files (images, videos) attached to diary entries via asset names.
- Body of the diary entry supports markdown and includes links to assets.
- User should be able to edit diary entry for selected date and assets for this date at the same page.
