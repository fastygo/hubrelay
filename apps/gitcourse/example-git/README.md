# Example GitCourse Repository

This repository is the first public GitCourse template.

## What it includes

- A Vite + React + TypeScript starter app
- `.course/course.json` with 7 lessons
- `.course/ci/verify.sh` to validate lesson progress
- `.course/progress.json` as the public student progress state
- `.github/workflows/course-check.yml` to update progress from CI
- `.cursor/rules/course.mdc` to guide the IDE agent

## Student workflow

1. Click **Use this template** or fork the repository.
2. Run `npm install`.
3. Start with `npm run dev`.
4. Complete the lessons in `.course/course.json`.
5. Push changes and let CI update `.course/progress.json`.
