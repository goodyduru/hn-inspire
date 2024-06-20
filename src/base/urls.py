"""
URL configuration for base project.

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/5.0/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  path('', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  path('', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.urls import include, path
    2. Add a URL to urlpatterns:  path('blog/', include('blog.urls'))
"""

from django.contrib import admin
from django.contrib.auth import views as auth_views
from django.urls import path

from member import views as member_views
from post import views as post_views

urlpatterns = [
    path("", post_views.IndexView.as_view(), name="index"),
    path("item/", post_views.PostView.as_view(), name="post"),
    path("comment/", post_views.CommentView.as_view(), name="comment"),
    path("submit/", post_views.SubmitView.as_view(), name="submit"),
    path("login/", member_views.LoginView.as_view(), name="login"),
    path("register/", member_views.RegisterView.as_view(), name="register"),
    path(
        "change-password/",
        auth_views.PasswordResetView.as_view(),
        name="password_reset",
    ),
    path("admin/", admin.site.urls),
]
