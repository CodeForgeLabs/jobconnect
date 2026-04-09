# User Service Profile Resource Endpoints

This document provides a detailed reference for authenticated self-service user endpoints exposed by the gateway under `/api/v1/users/me`.

## Common Conventions

### Authentication
- All endpoints in this document require authentication.
- Requests without a valid authenticated caller return `401 Unauthorized`.

### Request Size Limits
- JSON endpoints are limited to `1 MiB` request body size.
- When request body size is exceeded, gateway handlers return `413 Request Entity Too Large`.

### Pagination
- List endpoints accept:
  - `page_size`: integer in range `1..100`.
  - `page_token`: opaque cursor string.
- If omitted, `page_size` defaults to `20`.

### Error Mapping (High Level)
- `400 Bad Request`: validation failure, malformed payload, invalid path/query values.
- `401 Unauthorized`: missing/invalid authentication.
- `403 Forbidden`: permission denied (for example role-restricted operations).
- `404 Not Found`: target resource does not exist.
- `413 Request Entity Too Large`: body exceeds configured limit.
- `500 Internal Server Error`: unexpected backend/gateway error.

## Route Catalog

### Profile + Onboarding
1. `GET /api/v1/users/me/profile`
2. `PATCH /api/v1/users/me/profile`
3. `DELETE /api/v1/users/me/profile`
4. `GET /api/v1/users/me/onboarding-status`

### Account Settings
1. `GET /api/v1/users/me/settings`
2. `PATCH /api/v1/users/me/settings`

### Avatar (Direct Upload + Finalize)
1. `POST /api/v1/users/me/avatar/upload-url`
2. `POST /api/v1/users/me/avatar`
3. `GET /api/v1/users/me/avatar`
4. `DELETE /api/v1/users/me/avatar`

### CV (Direct Upload + Finalize)
1. `POST /api/v1/users/me/cv/upload-url`
2. `POST /api/v1/users/me/cv`
3. `GET /api/v1/users/me/cv`
4. `DELETE /api/v1/users/me/cv`

### Portfolio
1. `POST /api/v1/users/me/portfolio/media/upload-url`
2. `POST /api/v1/users/me/portfolio`
3. `GET /api/v1/users/me/portfolio`
4. `GET /api/v1/users/me/portfolio/:itemId`
5. `PUT /api/v1/users/me/portfolio/:itemId`
6. `DELETE /api/v1/users/me/portfolio/:itemId`

### Work + Hiring Preferences
1. `PATCH /api/v1/users/me/work-preferences`
2. `GET /api/v1/users/me/work-preferences`
3. `GET /api/v1/users/me/hiring-preferences`
4. `PATCH /api/v1/users/me/hiring-preferences`

### Saved Freelancers + Notes
1. `POST /api/v1/users/me/saved-freelancers/:freelancerId`
2. `GET /api/v1/users/me/saved-freelancers`
3. `DELETE /api/v1/users/me/saved-freelancers/:freelancerId`
4. `PUT /api/v1/users/me/freelancer-notes/:freelancerId`
5. `GET /api/v1/users/me/freelancer-notes/:freelancerId`

## Detailed Endpoint Reference

### Profile + Onboarding

#### GET /api/v1/users/me/profile
- Request body: none.
- Response `200`:
  - `profile`: `UserProfile`.
  - `completeness`: `ProfileCompleteness`.

#### PATCH /api/v1/users/me/profile
- Request body (supported fields):
  - Core: `display_name`, `contact_email`, `contact_phone`, `bio`, `tax_id`, `location`.
  - Client: `company_name`.
  - Freelancer: `headline`, `skills`, `hourly_rate`, `availability`.
- Response `200`:
  - `profile`: updated `UserProfile`.
  - `completeness`: updated `ProfileCompleteness`.
- Validation and behavior notes:
  - `avatar_url` is not patchable here. Use avatar endpoints.
  - `language` is not patchable here. Use `PATCH /api/v1/users/me/settings` with `ui_locale`.
  - `first_name`, `last_name`, and `last_active_at_unix` are unsupported in this endpoint.
  - `client` and `freelancer` fields cannot be patched together in one request.
  - At least one updatable field is required.
  - `tax_id` update is blocked when verification status is `submitted`, `pending_review`, or `verified`.

#### DELETE /api/v1/users/me/profile
- Query params:
  - `hard_delete` optional boolean.
- Response `200`:
  - `deleted`: boolean.

#### GET /api/v1/users/me/onboarding-status
- Request body: none.
- Response `200`:
  - `completeness`: `ProfileCompleteness`.
  - `readiness`: `ProfileReadiness`.
  - `steps`: `OnboardingStep[]`.

### Account Settings

#### GET /api/v1/users/me/settings
- Request body: none.
- Response `200`:
  - `settings`: `UserSettings`.

#### PATCH /api/v1/users/me/settings
- Request body (all optional, but at least one required):
  - `ui_locale`
  - `email_notifications_enabled`
  - `push_notifications_enabled`
- Response `200`:
  - `settings`: updated `UserSettings`.

### Avatar (Direct Upload + Finalize)

#### POST /api/v1/users/me/avatar/upload-url
- Request body:
  - `file_name`: string.
  - `content_type`: string.
- Response `200`:
  - `storage_key`: string.
  - `upload_url`: string (presigned upload URL).
- Notes:
  - Caller uploads binary directly to object storage using `upload_url`.

#### POST /api/v1/users/me/avatar
- Request body:
  - `storage_key`: string (required).
  - `file_name`: string.
  - `content_type`: string (required, must be supported image content type).
  - `width`: int32 optional.
  - `height`: int32 optional.
- Response `200`:
  - `avatar_url`: string.
  - `avatar`: `ProfileAvatar` (includes metadata and `download_url`).
- Validation and behavior notes:
  - `storage_key` is required.
  - Allowed `content_type`: `image/jpeg`, `image/png`, `image/webp`.

#### GET /api/v1/users/me/avatar
- Request body: none.
- Response `200`:
  - `avatar`: `ProfileAvatar`.
- Notes:
  - Current flow returns metadata including `download_url` for object retrieval.

#### DELETE /api/v1/users/me/avatar
- Request body: none.
- Response `200`:
  - `removed`: boolean.

### CV (Direct Upload + Finalize)

#### POST /api/v1/users/me/cv/upload-url
- Request body:
  - `file_name`: string.
  - `content_type`: string.
- Response `200`:
  - `storage_key`: string.
  - `upload_url`: string (presigned upload URL).

#### POST /api/v1/users/me/cv
- Request body:
  - `storage_key`: string (required).
  - `file_name`: string.
  - `content_type`: string (required).
- Response `200`:
  - `cv`: `ProfileCV` (includes `download_url`).
- Validation and behavior notes:
  - `storage_key` is required.
  - Allowed `content_type` values:
    - `application/pdf`
    - `application/msword`
    - `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
  - CV operations are role-restricted by user service policy.

#### GET /api/v1/users/me/cv
- Request body: none.
- Response `200`:
  - `cv`: `ProfileCV`.

#### DELETE /api/v1/users/me/cv
- Request body: none.
- Response `200`:
  - `removed`: boolean.

### Portfolio

#### POST /api/v1/users/me/portfolio/media/upload-url
- Request body:
  - `file_name`: string.
  - `content_type`: string.
- Response `200`:
  - `storage_key`: string.
  - `upload_url`: string (presigned upload URL).
- Notes:
  - Freelancer role is required (`403 Forbidden` for non-freelancer callers).
  - Caller uploads binary directly to object storage using `upload_url`.
  - Use the returned `storage_key` in a portfolio media item (`media[].storage_key`) when calling create/update endpoints.
  - Allowed `content_type`: `image/jpeg`, `image/png`, `image/webp`, `video/mp4`, `video/webm`, `application/pdf`.

#### POST /api/v1/users/me/portfolio
- Request body:
  - `title` (max 50 chars)
  - `description` (max 200 chars)
  - `project_url`
  - `role_in_project`
  - `completed_at_unix` optional
  - `tags` string list
  - `media` array (`PortfolioMediaInput`)
- Response `200`:
  - `item`: `PortfolioItem`.
- Notes:
  - Authenticated caller identity is always used for `user_id`.
  - For IMAGE/VIDEO/FILE media entries, request upload URL first and pass the returned `storage_key` in `media[].storage_key`.
  - For IMAGE/VIDEO/FILE media entries:
    - `storage_key` must start with `portfolio/{caller_user_id}/`.
    - `content_type` is required and must match media type:
      - IMAGE: `image/jpeg`, `image/png`, `image/webp`
      - VIDEO: `video/mp4`, `video/webm`
      - FILE: `application/pdf`

#### GET /api/v1/users/me/portfolio
- Query params:
  - `page_size`
  - `page_token`
- Response `200`:
  - `items`: `PortfolioItem[]`.
  - `next_page_token`: string.

#### GET /api/v1/users/me/portfolio/:itemId
- Path params:
  - `itemId`: positive integer (required).
- Response `200`:
  - `item`: `PortfolioItem`.

#### PUT /api/v1/users/me/portfolio/:itemId
- Path params:
  - `itemId`: positive integer (required).
- Request body:
  - Partial update payload matching `UpdateMyPortfolioItemRequest` fields.
- Response `200`:
  - `item`: updated `PortfolioItem`.
- Notes:
  - Authenticated caller `user_id` and path `itemId` override request body values.
  - For IMAGE/VIDEO/FILE media entries in updates:
    - `storage_key` must start with `portfolio/{caller_user_id}/`.
    - `content_type` is required and must match media type:
      - IMAGE: `image/jpeg`, `image/png`, `image/webp`
      - VIDEO: `video/mp4`, `video/webm`
      - FILE: `application/pdf`

#### DELETE /api/v1/users/me/portfolio/:itemId
- Path params:
  - `itemId`: positive integer (required).
- Response `200`:
  - `deleted`: boolean.

### Work + Hiring Preferences

#### PATCH /api/v1/users/me/work-preferences
- Request body (`PatchMyWorkPreferencesRequest` shape):
  - `preferred_project_length` optional
  - `min_budget` optional
  - `max_budget` optional
  - `contract_types` optional string list
  - `weekly_capacity_hours` optional
- Response `200`:
  - `settings`: `WorkPreferences`.

#### GET /api/v1/users/me/work-preferences
- Request body: none.
- Response `200`:
  - `settings`: `WorkPreferences`.

#### GET /api/v1/users/me/hiring-preferences
- Request body: none.
- Response `200`:
  - `preferences`: `HiringPreferences`.

#### PATCH /api/v1/users/me/hiring-preferences
- Request body (`PatchMyHiringPreferencesRequest` shape):
  - `min_hourly_rate` optional
  - `max_hourly_rate` optional
  - `preferred_locations` optional string list
- Response `200`:
  - `preferences`: updated `HiringPreferences`.

### Saved Freelancers + Notes

#### POST /api/v1/users/me/saved-freelancers/:freelancerId
- Path params:
  - `freelancerId`: string (required).
- Response `200`:
  - `saved`: `SavedFreelancer`.

#### GET /api/v1/users/me/saved-freelancers
- Query params:
  - `page_size`
  - `page_token`
- Response `200`:
  - `freelancers`: `SavedFreelancer[]`.
  - `next_page_token`: string.

#### DELETE /api/v1/users/me/saved-freelancers/:freelancerId
- Path params:
  - `freelancerId`: string (required).
- Response `200`:
  - `removed`: boolean.

#### PUT /api/v1/users/me/freelancer-notes/:freelancerId
- Path params:
  - `freelancerId`: string (required).
- Request body:
  - `note`: string.
- Response `200`:
  - `note`: `FreelancerNote`.
- Notes:
  - Authenticated caller `user_id` and path `freelancerId` override body values.

#### GET /api/v1/users/me/freelancer-notes/:freelancerId
- Path params:
  - `freelancerId`: string (required).
- Response `200`:
  - `note`: `FreelancerNote`.
