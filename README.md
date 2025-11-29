# 🤖 Robot Secret Reconstruction Simulation

Ce projet simule la collaboration de **plusieurs robots** pour reconstruire un secret partagé sous forme de mots.  
Chaque robot reçoit un sous-ensemble du secret et échange des messages par **canaux Go** jusqu'à ce que l’un d’eux recompose toute la phrase.

La simulation inclut :
- Perte et duplication aléatoire de messages
- Exécution concurrente via goroutines
- Propagation progressive des mots
- Arrêt propre lorsque le secret est reconstruit
- Écriture du secret final dans un fichier unique

---

# 📦 Génération du code Protobuf

Les fichiers `.proto` se trouvent dans le dossier **/proto**.  
Pour générer le code Go correspondant **sur Windows, Linux ou macOS**, lance simplement :

> **IMPORTANT :** La commande doit être exécutée *à la racine du projet.

```bash
docker run --rm -v "${PWD}/proto:/defs" namely/protoc-go ls /defs/proto