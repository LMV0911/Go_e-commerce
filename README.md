<div align="center">
  <img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" alt="Golang Logo" width="150" />
  <h1>Projet E-Commerce CLI en Go</h1>
  <p>Plateforme de vente en ligne complete et distribuee, accessible integralement en ligne de commande (CLI et SSH).</p>
</div>

---

## Architecture et Technologies

L'application respecte les contraintes techniques du sujet sans framework tiers :
* **Serveur HTTP central** : Go standard (`net/http` natif avec routage Go 1.22+).
* **Persistance SQL** : PostgreSQL avec le driver standard `database/sql` (sans ORM).
* **Interfaces Terminal Reactives (TUI)** : Ecosysteme **Charmbracelet** (`bubbletea`, `lipgloss`).
* **Serveur SSH distant** : **Charm Wish** sur le port 2222 pour administrer et commander a distance.

```text
Go_e-commerce/
├── cmd/
│   ├── server/          # Serveur HTTP principal
│   ├── admin/           # CLI Terminal Administrateur
│   ├── client/          # CLI Terminal Client
│   └── ssh/             # Serveur SSH Wish distant (port 2222)
├── internal/
│   ├── auth/            # Hachage bcrypt et signatures JWT
│   ├── database/        # Connexion PostgreSQL database/sql
│   ├── handler/         # Controleurs HTTP net/http (Admin, Auth, Client, Router)
│   ├── middleware/      # Middleware HTTP de verification JWT
│   ├── models/          # Structs de donnees typees
│   ├── repository/      # Requetes SQL et transactions atomiques
│   ├── tui/
│   │   ├── admin/       # Interface Bubbletea Admin avec sous-menus et confirmations
│   │   └── client/      # Interface Bubbletea Client avec panier et paiement
│   └── utils/           # Generateurs d'identifiants metier (PDT-, BSK-, CMD-)
├── migrations/
│   └── init.sql         # Schema relationnel, index, contraintes et seed data
├── Dockerfile           # Multi-stage build Go 1.24 / Alpine
└── docker-compose.yml   # Orchestration BDD, Adminer, HTTP Server & SSH
```

---

## Fonctionnalites

### Espace Client
* **Catalogue dynamique** : Recherche instantanee par mot-cle, filtre cyclique par categorie (`[c]`), tri par prix/nom (`[s]`), bascule Prix HT / Prix TTC (`[t]`).
* **Panier d'achat (`BSK-XXXXXX`)** : Ajout d'articles, modification de quantite, suppression d'article ou vidage complet. Les frais de livraison sont offerts (0.00 EUR).
* **Paiement et Commandes (`CMD-XXXXXX`)** : Formulaire de carte bancaire (16 chiffres, date d'expiration, CVC). Transaction SQL atomique avec decrement du stock.

### Espace Administrateur
* **Gestion des Produits (`PDT-XXXXXX`)** : Creation avec calcul automatique de la TVA (20%), mise a jour avec pre-remplissage des champs existants, suppression, et renommage de categorie insensible aux accents avec l'extension `unaccent`.
* **Gestion des Utilisateurs** : Listing des comptes, confirmation manuelle immediate, creation directe avec hachage `bcrypt`, mise a jour et suppression.
* **Gestion des Commandes** : Consultation du detail des articles et quantites, modification de statut avec motif d'annulation obligatoire.
* **Reintegration de stock automatique** : L'annulation d'une commande restitue instantanement les articles dans l'inventaire au sein d'une transaction SQL ACID.
* **Ergonomie TUI** : Ecrans de resultats dedies plein ecran sur 100% des actions et gestion UTF-8 par runes (`deleteLastRune`) pour eviter la corruption de caracteres accentues.

### Authentification et Securite
* Inscription avec mot de passe hache `bcrypt` (cout 10) et generation d'un code unique aleatoire a 6 chiffres.
* Validation de compte par code et reinitialisation de mot de passe oublie.
* Authentification JWT avec validation par middleware HTTP `RequireAuth`.

---

## Installation et Demarrage

### 1. Cloner le depot
```bash
git clone https://github.com/LMV0911/Go_e-commerce.git
cd Go_e-commerce
```

### 2. Configurer l'environnement
Copiez le fichier d'exemple pour initialiser la configuration locale :
```bash
cp .env.example .env
```
Editez le fichier `.env` pour renseigner imperativement les variables sensibles :
* `POSTGRES_PASSWORD` : Definissez le mot de passe de la base de donnees (ex: `password`).
* `JWT_SECRET` : Definissez la cle secrete de signature des tokens JWT (ex: `mon_secret_jwt_super_securise_123`).

### 3. Demarrer les conteneurs Docker
```bash
docker compose up -d
```
Les services suivants sont automatiquement demarres :
* **Serveur HTTP** : `http://localhost:8080`
* **Serveur SSH Wish** : `localhost:2222`
* **Base PostgreSQL** : `localhost:5432`
* **Adminer Web** : `http://localhost:8081`

---

## Utilisation des Interfaces CLI

### Lancer le TUI Client
```bash
# Via Docker :
docker compose run --rm client /app/bin/client

# Ou directement en local :
go run ./cmd/client
```

### Lancer le TUI Administrateur
```bash
# Via Docker :
docker compose run --rm admin /app/bin/admin

# Ou directement en local :
go run ./cmd/admin
```

### Acces Distant SSH (Wish - Port 2222)
Connexion directe depuis n'importe quel terminal sans installation prealable :
```bash
# Interface Client :
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null client@localhost -p 2222

# Interface Administration :
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null admin@localhost -p 2222
```

> **Identifiants Admin par defaut** :  
> Email : `admin@ecommerce.com` | Mot de passe : `admin123`

---

## Documentation des Endpoints HTTP

### Authentification (`/api/auth`)
| Methode | Route | Description | Protection |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/auth/register` | Inscription client (hachage bcrypt + code 6 chiffres) | Publique |
| `POST` | `/api/auth/confirm` | Confirmation d'un compte via code unique | Publique |
| `POST` | `/api/auth/login` | Connexion utilisateur et emission du JWT | Publique |
| `POST` | `/api/auth/forgot-password` | Demande de code de reinitialisation par email | Publique |
| `POST` | `/api/auth/reset-password` | Reinitialisation du mot de passe avec le code | Publique |
| `GET` | `/api/auth/me` | Profil utilisateur connecte | **Bearer JWT** |

### Espace Client (`/api/...`)
| Methode | Route | Description |
| :--- | :--- | :--- |
| `GET` | `/api/products` | Catalogue avec filtres (`?q=`, `?category=`, `?sort=`, `?price_type=`) |
| `GET` | `/api/categories` | Liste dynamique des categories ayant des produits en stock |
| `GET` | `/api/cart` | Recuperation du panier en cours (`BSK-XXXXXX`) |
| `POST` | `/api/cart/items` | Ajout d'un produit au panier |
| `PUT` | `/api/cart/items` | Modification de la quantite d'un article |
| `DELETE` | `/api/cart/items` | Suppression d'un article du panier |
| `DELETE` | `/api/cart` | Vidage complet du panier |
| `POST` | `/api/checkout` | Validation de commande (`CMD-XXXXXX`) et paiement CB |
| `GET` | `/api/orders` | Historique des commandes du client connecte |

### Espace Administration (`/api/admin/...`)
| Methode | Route | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Healthcheck de supervision |
| `GET` | `/api/admin/products` | Liste de tous les produits du catalogue |
| `GET` | `/api/admin/products/{id}` | Detail d'un produit (`PDT-XXXXXX`) |
| `POST` | `/api/admin/products` | Creation de produit (ID automatique, calcul TVA 20%) |
| `PUT` | `/api/admin/products/{id}` | Modification des attributs d'un produit |
| `DELETE` | `/api/admin/products/{id}` | Suppression d'un produit |
| `PUT` | `/api/admin/categories/rename` | Renommage d'une categorie insensible aux accents (`unaccent`) |
| `GET` | `/api/admin/users` | Liste des utilisateurs enregistres |
| `GET` | `/api/admin/users/{id}` | Detail d'un compte utilisateur |
| `POST` | `/api/admin/users` | Creation manuelle d'un compte utilisateur par l'admin |
| `POST` | `/api/admin/users/{id}/confirm` | Validation / activation manuelle d'un compte |
| `PUT` | `/api/admin/users/{id}` | Mise a jour des informations utilisateur |
| `DELETE` | `/api/admin/users/{id}` | Suppression d'un compte utilisateur |
| `GET` | `/api/admin/orders` | Liste globale des commandes avec jointure client |
| `GET` | `/api/admin/orders/{id}` | Detail d'une commande avec liste des articles |
| `PUT` | `/api/admin/orders/{id}/status` | Changement de statut (motif obligatoire et remise en stock si annule) |
| `POST` | `/api/admin/orders/manual` | Creation d'une commande manuelle pour un compte client tiers |

---

## Tests du Projet
Pour executer l'ensemble de la suite de tests automatisee :
```bash
docker compose run --rm server go test -v ./...
# Ou en local :
go test -v ./...
```
