-- Move all PEMIRA tables from public schema to myschema schema

-- Create myschema if not exists
CREATE SCHEMA IF NOT EXISTS myschema;

-- Move tables to myschema
ALTER TABLE IF EXISTS public.app_settings SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.branding_files SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.branding_settings SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.candidate_media SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.candidate_qr_codes SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.candidates SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.election_voters SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.elections SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.faculties SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.lecturer_positions SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.lecturer_units SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.lecturers SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.migration_history SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.registration_tokens SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.schema_migrations SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.staff_members SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.staff_positions SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.staff_units SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.students SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.study_programs SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.tps SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.tps_checkins SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.tps_pic SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.user_accounts SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.voters SET SCHEMA myschema;
ALTER TABLE IF EXISTS public.votes SET SCHEMA myschema;

-- Move sequences to myschema
ALTER SEQUENCE IF EXISTS public.app_settings_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.branding_files_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.branding_settings_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.candidate_media_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.candidate_qr_codes_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.candidates_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.election_voters_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.elections_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.faculties_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.lecturer_positions_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.lecturer_units_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.lecturers_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.migration_history_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.registration_tokens_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.staff_members_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.staff_positions_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.staff_units_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.students_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.study_programs_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.tps_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.tps_checkins_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.tps_pic_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.user_accounts_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.voters_id_seq SET SCHEMA myschema;
ALTER SEQUENCE IF EXISTS public.votes_id_seq SET SCHEMA myschema;

-- Verify the move
SELECT 'Tables in myschema:' AS info;
SELECT table_name FROM information_schema.tables WHERE table_schema = 'myschema' ORDER BY table_name;

SELECT 'Row counts:' AS info;
SELECT 'elections' AS table_name, COUNT(*) AS rows FROM myschema.elections
UNION ALL
SELECT 'voters', COUNT(*) FROM myschema.voters
UNION ALL
SELECT 'candidates', COUNT(*) FROM myschema.candidates
UNION ALL
SELECT 'votes', COUNT(*) FROM myschema.votes
UNION ALL
SELECT 'tps', COUNT(*) FROM myschema.tps
UNION ALL
SELECT 'user_accounts', COUNT(*) FROM myschema.user_accounts;
