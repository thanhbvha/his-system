DELETE FROM permissions 
WHERE resource IN ('users', 'roles', 'departments', 'patients', 'appointments', 'billing');
