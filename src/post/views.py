from django.contrib.auth import get_user_model
from django.contrib.auth.mixins import LoginRequiredMixin
from django.http import Http404, HttpResponseBadRequest
from django.shortcuts import render, redirect
from django.views import View

from .forms import AddComment, AddPost
from .models import Post
from .sql import get_home_page, get_comments, get_new_posts


# Create your views here.
class IndexView(View):
    template_name = "post/index.html"
    title = "home"
    header = "Home"

    def get(self, request, *args, **kwargs):
        latest_posts = get_home_page(request.user)
        context = {
            "latest_post_list": latest_posts,
            "title": self.title,
            "header": self.header,
        }
        return render(request, template_name=self.template_name, context=context)


class NewPostView(View):
    template_name = "post/index.html"
    title = "new"
    header = "New Posts"

    def get(self, request, *args, **kwargs):
        new_posts = get_new_posts(request.user)
        context = {
            "latest_post_list": new_posts,
            "title": self.title,
            "header": self.header,
        }
        return render(request, template_name=self.template_name, context=context)


class PostView(View):
    template_name = "post/detail.html"
    comment_form = AddComment

    def get(self, request, *args, **kwargs):
        post_id = request.GET.get("id")
        error404 = Http404("Item does not exist")
        if post_id is None:
            raise error404
        posts = get_comments(request.user, int(post_id))
        if len(posts) == 0:
            raise error404
        post = posts[0]
        comments = posts[1:]
        form = self.comment_form(author=request.user, initial={"parent": post.id})
        return render(
            request,
            self.template_name,
            {
                "post": post,
                "comment_list": comments,
                "form": form,
            },
        )


class CommentView(LoginRequiredMixin, View):
    comment_form = AddComment

    def post(self, request, *args, **kwargs):
        comment_form = self.comment_form(data=request.POST, author=request.user)
        if comment_form.is_valid():
            post = comment_form.save()
            goto = request.GET.get("goto", f"{post.parent.pk}#{post.pk}")
            redirect_url = f"/item?id={goto}"
            return redirect(redirect_url)
        else:
            return HttpResponseBadRequest(content="Form contains invalid data")


class ReplyView(LoginRequiredMixin, View):
    template_name = "post/reply.html"
    comment_form = AddComment

    def get(self, request, *args, **kwargs):
        id = request.GET.get("id")
        goto = request.GET.get("goto", "")
        try:
            post = Post.objects.select_related("author").get(pk=int(id))
        except:
            raise Http404("Item does not exist")
        if post.parent is None:
            return HttpResponseBadRequest(content="Bad request")
        form = self.comment_form(author=request.user, initial={"parent": post.id})
        return render(
            request, self.template_name, {"post": post, "goto": goto, "form": form}
        )


class SubmitView(LoginRequiredMixin, View):
    submit_form = AddPost
    template_name = "post/submit.html"

    def get(self, request, *args, **kwargs):
        submit_form = self.submit_form(author=request.user)
        return render(request, self.template_name, {"form": submit_form})

    def post(self, request, *args, **kwargs):
        submit_form = self.submit_form(data=request.POST, author=request.user)
        if submit_form.is_valid():
            _ = submit_form.save()
            return redirect("index")
        else:
            return render(request, self.template_name, {"form": submit_form})


class SubmittedView(View):
    template_name = "post/index.html"

    def get(self, request, *args, **kwargs):
        username = request.GET.get("id")
        try:
            user = get_user_model().objects.get(username=username)
        except:
            raise Http404("No such user exists")
        submitted_posts = (
            Post.objects.filter(parent=None, author=user)
            .select_related("author")
            .order_by("-updated_at")[:5]
        )
        context = {"latest_post_list": submitted_posts, "header": f"{username} Posts"}
        return render(request, template_name=self.template_name, context=context)


class ThreadView(View):
    template_name = "post/thread.html"

    def get(self, request, *args, **kwargs):
        username = request.GET.get("id")
        try:
            user = get_user_model().objects.get(username=username)
        except:
            raise Http404("No such user exists")
        comments = (
            Post.objects.filter(parent__isnull=False, author=user)
            .select_related("author")
            .order_by("-updated_at")[:5]
        )
        context = {"comments": comments}
        return render(request, template_name=self.template_name, context=context)


class FavoritesView(View):
    template_name = "post/index.html"

    def get(self, request, *args, **kwargs):
        username = request.GET.get("id")
        try:
            user = get_user_model().objects.get(username=username)
        except:
            raise Http404("No such user exists")
        submitted_posts = (
            Post.objects.filter(parent=None, favoriters=user)
            .select_related("author")
            .order_by("-updated_at")[:5]
        )
        context = {
            "latest_post_list": submitted_posts,
            "header": f"{username} Favorites",
        }
        return render(request, template_name=self.template_name, context=context)


class FavePostView(LoginRequiredMixin, View):
    def get(self, request, *args, **kwargs):
        id = request.GET.get("id")
        action = request.GET.get("action")
        error404 = Http404("Item does not exist")
        if id is None or action not in [None, "un"]:
            raise error404
        try:
            post = Post.objects.select_related("author").get(pk=int(id))
        except:
            raise error404
        user = request.user
        if user.is_authenticated and post.author.username == user.username:
            return HttpResponseBadRequest(content="Bad request")
        if action is None:
            post.favoriters.add(user)
        else:
            post.favoriters.remove(user)
        url = f"/item?id={id}"
        return redirect(url)


class VotePostView(LoginRequiredMixin, View):
    def get(self, request, *args, **kwargs):
        id = request.GET.get("id")
        action = request.GET.get("action")
        error404 = Http404("Item does not exist")
        if id is None or action not in [None, "un"]:
            raise error404
        try:
            post = Post.objects.prefetch_related("voters").get(pk=int(id))
        except:
            raise error404
        user = request.user
        if user.is_authenticated and post.author.username == user.username:
            return HttpResponseBadRequest(content="Bad request")
        count = post.voters.filter(id=user.pk).count()
        if count == 0 and action is None:
            post.voters.add(user)
            post.votes += 1
            post.save()
        elif count == 1 and action == "un":
            post.voters.remove(user)
            post.votes -= 1
            post.save()

        url = f"/item?id={id}"
        return redirect(url)


class FlagPostView(LoginRequiredMixin, View):
    def get(self, request, *args, **kwargs):
        id = request.GET.get("id")
        action = request.GET.get("action")
        error404 = Http404("Item does not exist")
        if id is None or action not in [None, "un"]:
            raise error404
        try:
            post = Post.objects.get(pk=int(id))
        except:
            raise error404
        user = request.user
        if user.is_authenticated and post.author.username == user.username:
            return HttpResponseBadRequest(content="Bad request")
        if action is None:
            post.flaggers.add(user)
        else:
            post.flaggers.remove(user)
        url = f"/item?id={id}"
        return redirect(url)
