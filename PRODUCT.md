# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- Patients book, reschedule, and cancel visits, and look up upcoming and past appointments.
- Doctors review their daily agenda and weekly calendar, set working hours and time blocks, and manage visits within their permissions.
- Clinic administrators and front-desk staff coordinate the clinic diary, book on behalf of patients, and manage schedules and accounts within their permissions.

Audience priority when these workflows compete is open. The roles and jobs above are drawn from the existing application and repository documentation.

## Product Purpose

CuraClinic gives patients and clinic teams a shared view of appointment availability. Patients can find a doctor and confirm a visit while staff coordinate the clinic diary and clinicians keep their schedules current.

## Operating Context

Patients see availability based on a doctor's working hours, blocked time, confirmed visits, and active holds. Choosing a time temporarily reserves it during self-service booking or rescheduling. Patients see a countdown and can choose another time if the hold expires. Clinicians and staff with permission can book manually for a patient.

## Capabilities and Constraints

- Patients sign up for their own accounts. After sign-in, the application opens a role-specific portal and uses effective permissions to control which routes and actions are available.
- Booking rules prevent overlapping confirmed appointments for both doctors and patients. Cancelling an appointment releases its slot immediately.
- Appointment history records reschedules and cancellations. The clinic audit trail records sensitive changes such as bookings, account changes, and access changes.

## Evidence on Hand

- [README.md](README.md) describes the three portals, booking rules, and demo setup.
- [The app router](frontend/src/app-router.tsx) shows the available routes and permission checks.
- [The sign-in and sign-up layout](frontend/src/components/auth/auth-layout.tsx) contains the current CuraClinic branding and public-facing product copy.

## Product Principles

These principles are inferred from the implemented workflows and are pending confirmation.

- Show current availability and make booking state clear at each step.
- Give each role the next action it needs within its permissions.
- Keep appointment changes traceable across the three portals.
