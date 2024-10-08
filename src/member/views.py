from django.conf import settings
from django.contrib.auth import login, logout
from django.contrib.auth.decorators import login_required
from django.contrib.auth.forms import AuthenticationForm
from django.contrib.auth import get_user_model
from django.contrib.auth.views import RedirectURLMixin
from django.http import HttpResponseRedirect, Http404, HttpResponseForbidden
from django.shortcuts import render, resolve_url
from django.utils.decorators import method_decorator
from django.views import View

from .forms import AddMemberForm, UpdateMemberForm


# Create your views here.
class LoginView(View, RedirectURLMixin):
    login_form_class = AuthenticationForm
    register_form_class = AddMemberForm
    template_name = "member/login.html"

    def get(self, request, *args, **kwargs):
        login_form = self.login_form_class()
        register_form = self.register_form_class()
        next = request.GET.get("next", "/")
        return render(
            request,
            self.template_name,
            {"login_form": login_form, "register_form": register_form, "next": next},
        )

    def post(self, request, *args, **kwargs):
        login_form = self.login_form_class(data=request.POST)
        register_form = self.register_form_class()
        next = request.GET.get("next")
        if login_form.is_valid():
            login(request, login_form.get_user())
            return HttpResponseRedirect(self.get_success_url())
        else:
            return render(
                request,
                self.template_name,
                {
                    "login_form": login_form,
                    "register_form": register_form,
                    "next": next,
                },
            )

    def get_default_redirect_url(self):
        """Return the default redirect URL."""
        return resolve_url(settings.LOGIN_REDIRECT_URL)


class RegisterView(View, RedirectURLMixin):
    login_form_class = AuthenticationForm
    register_form_class = AddMemberForm
    template_name = "member/login.html"

    def get(self, request, *args, **kwargs):
        return HttpResponseRedirect(resolve_url("login"))

    def post(self, request, *args, **kwargs):
        login_form = self.login_form_class()
        register_form = self.register_form_class(request.POST)
        next = request.GET.get("next")
        if register_form.is_valid():
            user = register_form.save()
            login(request, user)
            return HttpResponseRedirect(self.get_success_url())
        else:
            return render(
                request,
                self.template_name,
                {
                    "login_form": login_form,
                    "register_form": register_form,
                    "next": next,
                },
            )

    def get_default_redirect_url(self):
        """Return the default redirect URL."""
        return resolve_url(settings.LOGIN_REDIRECT_URL)


class LogoutView(View):
    def get(self, request, *args, **kwargs):
        logout(request)
        return HttpResponseRedirect(resolve_url(settings.LOGOUT_REDIRECT_URL))


class ProfileView(View):
    template_name = "member/profile.html"
    form = UpdateMemberForm

    def get(self, request, *args, **kwargs):
        username = request.GET.get("id")
        try:
            member = (
                get_user_model()
                .objects.select_related("profile")
                .get(username=username)
            )
        except:
            raise Http404("No such user exists")
        if username == request.user.username:
            initial = {"email": member.email, "about": member.profile.about}
            form = self.form(member=member, initial=initial)
            return render(
                request,
                self.template_name,
                {"member": member, "form": form, "title": username},
            )
        else:
            return render(request, self.template_name, {"member": member})

    @method_decorator(login_required())
    def post(self, request, *args, **kwargs):
        username = request.GET.get("id")
        try:
            member = (
                get_user_model()
                .objects.select_related("profile")
                .get(username=username)
            )
        except:
            raise Http404("No such user exists")
        if username != request.user.username:
            raise HttpResponseForbidden(content="You have no permission to do that")
        initial = {"email": member.email, "about": member.profile.about}
        form = self.form(member=member, initial=initial, data=request.POST)
        if form.is_valid():
            member = form.save()
        return render(
            request,
            self.template_name,
            {"member": member, "form": form, "title": username},
        )
