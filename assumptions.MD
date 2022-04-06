# Choix techniques & discussions

## Performances
### Utilisation d'un GITHUB_TOKEN individuel
l'API Github à un limit rate de 5000 calls/heure par Utilisateur authentifié. L'utilisation d'un Token individuel est nécessaire pour un usage multi-utilisateurs. Si un utilisateur atteint son quota, les autres utilisateurs ne seront pas impactés.

### Parallélisation des Calls API pour récupérer les Langages de chaque repository.
Le Endpoint d'API GitHub de recherche ne permet pas de récupérer les informations relatives aux langages utilisés. Il faut utiliser pour chaque Repository, un endpoint spécifique <repo>/languages. Non parallélisé, le temps de réponse de l'API est délirant (env. 8-9s).

Utilisation du `sync.WaitGroup{}` pour paralléliser les call APIs. Le temps de réponse de l'API passe sous la barre des 2s (Toujours lent mais mieux)

### Piste(s) d'amélioration(s)
Utilisation d'un "In Memory Cache System" pour stocker les "Items" (Redis ou autre). De cette manière, on peut mutualiser les résultats entre les utilisateurs et limiter fortement les calls d'API
[https://adityarama1210.medium.com/fast-golang-api-performance-with-in-memory-key-value-storing-cache-1b248c182bdb](https://adityarama1210.medium.com/fast-golang-api-performance-with-in-memory-key-value-storing-cache-1b248c182bdb)

A tester ce soir...

## Maintenabilité dans le temps
### Tests
Ajouter des tests automatisés

### CI
Ajouter une pipeline de déploiement (Github Action ou Gitlab Pipeline).

