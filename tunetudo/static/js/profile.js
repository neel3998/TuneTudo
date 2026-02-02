// static/js/profile.js

// Check authentication
if (!isAuthenticated()) {
    window.location.href = 'index.html';
}

// Load profile data
async function loadProfile() {
    try {
        const response = await AuthAPI.getProfile();
        const user = response.data;
        
        displayProfile(user);
        loadStatistics();
    } catch (error) {
        showAlert('Failed to load profile', 'error');
    }
}

function displayProfile(user) {
    // Update profile info
    document.getElementById('profileUsername').textContent = user.username;
    document.getElementById('profileEmail').textContent = user.email;
    document.getElementById('profileRole').textContent = user.is_admin ? 'Administrator' : 'User';
    document.getElementById('profileCreatedAt').textContent = new Date(user.created_at).toLocaleDateString();
    
    if (user.last_login) {
        document.getElementById('profileLastLogin').textContent = new Date(user.last_login).toLocaleString();
    } else {
        document.getElementById('profileLastLogin').textContent = 'Never';
    }
    
    // Update profile picture
    if (user.profile_image_path) {
        document.getElementById('profileImage').src = `/storage/${user.profile_image_path}`;
    }
}

async function loadStatistics() {
    try {
        // Load playlists count
        const playlistsResponse = await PlaylistAPI.getUserPlaylists();
        const playlists = playlistsResponse.data;
        document.getElementById('playlistCount').textContent = playlists.length;
        
        // Calculate total songs in playlists
        let totalSongs = 0;
        playlists.forEach(playlist => {
            totalSongs += playlist.song_count || 0;
        });
        document.getElementById('totalSongsInPlaylists').textContent = totalSongs;
        
        // Load uploads count
        const uploadsResponse = await UserAPI.getUserUploads();
        document.getElementById('uploadCount').textContent = uploadsResponse.data.length;
    } catch (error) {
        console.error('Failed to load statistics:', error);
    }
}

// Upload picture modal
function showUploadPictureModal() {
    document.getElementById('uploadPictureModal').classList.add('show');
}

function closeUploadPictureModal() {
    document.getElementById('uploadPictureModal').classList.remove('show');
    document.getElementById('uploadPictureForm').reset();
    document.getElementById('imagePreview').style.display = 'none';
    document.getElementById('uploadProgress').style.display = 'none';
}

// Preview image before upload
document.addEventListener('DOMContentLoaded', () => {
    const pictureInput = document.getElementById('pictureFile');
    if (pictureInput) {
        pictureInput.addEventListener('change', function(e) {
            const file = e.target.files[0];
            if (file) {
                const reader = new FileReader();
                reader.onload = function(e) {
                    document.getElementById('previewImage').src = e.target.result;
                    document.getElementById('imagePreview').style.display = 'block';
                };
                reader.readAsDataURL(file);
            }
        });
    }
});

async function uploadProfilePicture(event) {
    event.preventDefault();
    
    const fileInput = document.getElementById('pictureFile');
    const file = fileInput.files[0];
    
    if (!file) {
        showAlert('Please select a file', 'error');
        return;
    }
    
    // Validate file size (5MB)
    if (file.size > 5 * 1024 * 1024) {
        showAlert('File too large. Maximum size is 5MB', 'error');
        return;
    }
    
    // Validate file type
    const validTypes = ['image/jpeg', 'image/jpg', 'image/png'];
    if (!validTypes.includes(file.type)) {
        showAlert('Invalid file type. Only JPG and PNG allowed', 'error');
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
        }, 100);
        
        await UserAPI.uploadProfilePicture(file);
        
        clearInterval(interval);
        progressFill.style.width = '100%';
        progressText.textContent = 'Upload complete!';
        
        showAlert('Profile picture updated successfully!');
        closeUploadPictureModal();
        
        // Reload profile to show new picture
        setTimeout(() => {
            loadProfile();
        }, 1000);
    } catch (error) {
        showAlert(error.message, 'error');
    } finally {
        uploadBtn.disabled = false;
        setTimeout(() => {
            progressDiv.style.display = 'none';
        }, 2000);
    }
}

// Change password modal
function showChangePasswordModal() {
    document.getElementById('changePasswordModal').classList.add('show');
}

function closeChangePasswordModal() {
    document.getElementById('changePasswordModal').classList.remove('show');
    document.getElementById('changePasswordForm').reset();
}

async function changePassword(event) {
    event.preventDefault();
    
    const currentPassword = document.getElementById('currentPassword').value;
    const newPassword = document.getElementById('newPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;
    
    if (newPassword !== confirmPassword) {
        showAlert('New passwords do not match', 'error');
        return;
    }
    
    if (newPassword.length < 8) {
        showAlert('Password must be at least 8 characters', 'error');
        return;
    }
    
    showAlert('Password change feature coming soon!', 'error');
    // TODO: Implement password change API endpoint
    closeChangePasswordModal();
}

// Delete account
function confirmDeleteAccount() {
    if (confirm('Are you sure you want to delete your account? This action cannot be undone!')) {
        if (confirm('This will permanently delete all your playlists and uploads. Are you absolutely sure?')) {
            deleteAccount();
        }
    }
}

async function deleteAccount() {
    showAlert('Account deletion feature coming soon!', 'error');
    // TODO: Implement account deletion API endpoint
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadProfile();
});