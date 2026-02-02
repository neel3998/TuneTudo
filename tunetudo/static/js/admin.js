// static/js/admin.js

// Check admin authentication
if (!isAuthenticated() || !isAdmin()) {
    window.location.href = 'index.html';
}

let currentPage = 0;
const pageSize = 50;
let deleteSongId = null;
let categories = [];

// Tab switching
function switchTab(tabName) {
    // Update tab buttons
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });
    event.target.classList.add('active');
    
    // Update tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(tabName + 'Tab').classList.add('active');
    
    // Load data for the tab
    if (tabName === 'songs') {
        loadSongs();
    } else if (tabName === 'users') {
        loadUsers();
    }
}

// Load categories for dropdown
async function loadCategories() {
    try {
        const response = await SearchAPI.getCategories();
        categories = response.data;
        
        const select = document.getElementById('categorySelect');
        if (select) {
            select.innerHTML = '<option value="">Select category...</option>' +
                categories.map(cat => `<option value="${cat.id}">${cat.name}</option>`).join('');
        }
    } catch (error) {
        console.error('Failed to load categories:', error);
    }
}

// Load songs
async function loadSongs() {
    try {
        const response = await AdminAPI.getAllSongs(pageSize, currentPage * pageSize);
        displaySongs(response.data);
    } catch (error) {
        showAlert('Failed to load songs', 'error');
    }
}

function displaySongs(songs) {
    const tbody = document.getElementById('songsTableBody');
    
    if (!songs || songs.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="text-center">No songs found</td></tr>';
        document.getElementById('nextBtn').disabled = true;
        return;
    }
    
    tbody.innerHTML = songs.map(song => {
        const duration = formatDuration(song.duration_seconds);
        return `
            <tr>
                <td>${song.id}</td>
                <td>${song.title}</td>
                <td>${song.artist ? song.artist.name : 'Unknown'}</td>
                <td>${song.album ? song.album.title : '-'}</td>
                <td>${song.category ? song.category.name : '-'}</td>
                <td>${duration}</td>
                <td>
                    <button class="btn btn-danger" onclick="showDeleteModal(${song.id}, '${(song.title)}')">Delete</button>
                </td>
            </tr>
        `;
    }).join('');
    
    updatePagination();
    document.getElementById('nextBtn').disabled = songs.length < pageSize;
}

function formatDuration(seconds) {
    if (!seconds) return '--:--';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function updatePagination() {
    document.getElementById('pageInfo').textContent = `Page ${currentPage + 1}`;
    document.getElementById('prevBtn').disabled = currentPage === 0;
}

function loadNextPage() {
    currentPage++;
    loadSongs();
}

function loadPreviousPage() {
    if (currentPage > 0) {
        currentPage--;
        loadSongs();
    }
}

// Load users
async function loadUsers() {
    try {
        const response = await AdminAPI.getAllUsers();
        displayUsers(response.data);
    } catch (error) {
        showAlert('Failed to load users', 'error');
    }
}

function displayUsers(users) {
    const tbody = document.getElementById('usersTableBody');
    
    if (!users || users.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="text-center">No users found</td></tr>';
        return;
    }
    
    tbody.innerHTML = users.map(user => `
        <tr>
            <td>${user.id}</td>
            <td>${user.username}</td>
            <td>${user.email}</td>
            <td>${user.is_admin ? '<span style="color: var(--primary-color)">Admin</span>' : 'User'}</td>
            <td>${new Date(user.created_at).toLocaleDateString()}</td>
            <td>${user.last_login ? new Date(user.last_login).toLocaleDateString() : 'Never'}</td>
        </tr>
    `).join('');
}

// Add song modal
function showAddSongModal() {
    document.getElementById('addSongModal').classList.add('show');
}

function closeAddSongModal() {
    document.getElementById('addSongModal').classList.remove('show');
    document.getElementById('addSongForm').reset();
    document.getElementById('uploadProgress').style.display = 'none';
}

async function uploadSong(event) {
    event.preventDefault();
    
    const fileInput = document.getElementById('songFile');
    const file = fileInput.files[0];
    
    if (!file) {
        showAlert('Please select a file', 'error');
        return;
    }
    
    // Validate file size (50MB)
    if (file.size > 50 * 1024 * 1024) {
        showAlert('File too large. Maximum size is 50MB', 'error');
        return;
    }
    
    const title = document.getElementById('songTitle').value.trim();
    const artist = document.getElementById('artistName').value.trim();
    const album = document.getElementById('albumTitle').value.trim();
    const categoryId = parseInt(document.getElementById('categorySelect').value) || 0;
    const duration = parseInt(document.getElementById('duration').value) || 0;
    
    if (!title || !artist) {
        showAlert('Title and Artist are required', 'error');
        return;
    }
    
    const uploadBtn = document.getElementById('uploadBtn');
    const progressDiv = document.getElementById('uploadProgress');
    const progressFill = document.getElementById('progressBarFill');
    const progressText = document.getElementById('progressText');
    
    uploadBtn.disabled = true;
    progressDiv.style.display = 'block';
    progressFill.style.width = '0%';
    progressText.textContent = 'Uploading...';
    
    try {
        // Simulate progress
        let progress = 0;
        const interval = setInterval(() => {
            progress += 10;
            if (progress <= 90) {
                progressFill.style.width = progress + '%';
            }
        }, 200);
        
        await AdminAPI.uploadSong(file, title, artist, album, categoryId, duration);
        
        clearInterval(interval);
        progressFill.style.width = '100%';
        progressText.textContent = 'Upload complete!';
        
        showAlert('Song uploaded successfully!');
        closeAddSongModal();
        loadSongs();
    } catch (error) {
        showAlert(error.message, 'error');
    } finally {
        uploadBtn.disabled = false;
        setTimeout(() => {
            progressDiv.style.display = 'none';
        }, 2000);
    }
}

// Delete song
function showDeleteModal(songId, songTitle) {
    deleteSongId = songId;
    document.getElementById('deleteSongTitle').textContent = songTitle;
    document.getElementById('deleteModal').classList.add('show');
}

function closeDeleteModal() {
    document.getElementById('deleteModal').classList.remove('show');
    deleteSongId = null;
}

async function confirmDelete() {
    if (!deleteSongId) return;
    
    try {
        await AdminAPI.deleteSong(deleteSongId);
        showAlert('Song deleted successfully');
        closeDeleteModal();
        loadSongs();
    } catch (error) {
        showAlert(error.message, 'error');
    }
}


// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadCategories();
    loadSongs();
});