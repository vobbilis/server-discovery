-- Move tables from public to server_discovery schema
ALTER TABLE public.servers SET SCHEMA server_discovery;
ALTER TABLE public.server_details SET SCHEMA server_discovery;
ALTER TABLE public.server_metrics SET SCHEMA server_discovery;
ALTER TABLE public.server_ports SET SCHEMA server_discovery;
ALTER TABLE public.server_services SET SCHEMA server_discovery;

-- Move functions
ALTER FUNCTION public.update_updated_at_column() SET SCHEMA server_discovery; 